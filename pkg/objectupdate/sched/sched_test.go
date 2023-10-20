/*
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2023 Red Hat, Inc.
 */

package sched

import (
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"

	pluginconfig "github.com/k8stopologyawareschedwg/k8sschedulerconfig-api/scheduler-plugins/apis/config"
)

func TestRenderConfig(t *testing.T) {
	rendered, err := RenderConfig(configTemplate, "test-sched-name", 42*time.Second)
	if err != nil {
		t.Errorf("RenderConfig() failed: %v", err)
	}

	schedCfg, err := manifests.DecodeSchedulerConfigFromData([]byte(rendered))
	if err != nil {
		t.Errorf("failed to decode rendered data: %v", err)
	}

	schedProf, pluginConf := findKubeSchedulerProfileByName(schedCfg, schedulerPluginName)
	if schedProf == nil || pluginConf == nil {
		t.Errorf("no profile or plugin configuration found for %q", schedulerPluginName)
	}

	confObj := pluginConf.Args.DeepCopyObject()
	pluginCfg, ok := confObj.(*pluginconfig.NodeResourceTopologyMatchArgs)
	if !ok {
		t.Errorf("unsupported plugin config type: %T", confObj)
	}

	if schedProf.SchedulerName != "test-sched-name" {
		t.Errorf("unexpected rendered data: scheduler profile name: %q", schedProf.SchedulerName)
	}

	if pluginCfg.CacheResyncPeriodSeconds != int64(42) {
		t.Errorf("unexpected rendered data: resync period: %d", pluginCfg.CacheResyncPeriodSeconds)
	}
}

func TestSchedulerDeployment(t *testing.T) {
	type testCase struct {
		name               string
		pullIfNotPresent   bool
		ctrlPlaneAffinity  bool
		verbose            int
		expectedRenderedDp string
	}

	testCases := []testCase{
		{
			name:               "defaults",
			verbose:            4, // TODO: this *IS* the default - see pkg/commands/root.go - but how do we keep this in sync?
			ctrlPlaneAffinity:  true,
			expectedRenderedDp: expectedSchedDeploymentDefault,
		},
		{
			name:               "non-affine",
			verbose:            4, // TODO: this *IS* the default - see pkg/commands/root.go - but how do we keep this in sync?
			expectedRenderedDp: expectedSchedDeploymentNonAffine,
		},
		{
			name:               "extra-verbose",
			verbose:            6,
			ctrlPlaneAffinity:  true,
			expectedRenderedDp: expectedSchedDeploymentVerbose,
		},
	}

	dpRef, err := manifests.Deployment(manifests.ComponentSchedulerPlugin, manifests.SubComponentSchedulerPluginScheduler, "")
	if err != nil {
		t.Errorf("cannot load the scheduler manifest: %v", err)
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dp := dpRef.DeepCopy()

			SchedulerDeployment(dp, tc.pullIfNotPresent, tc.ctrlPlaneAffinity, tc.verbose)
			fixSchedulerImage(dp)

			var sb strings.Builder
			manifests.RenderObjects([]client.Object{dp}, &sb)
			got := sb.String()

			if got != tc.expectedRenderedDp {
				t.Errorf("unexpected result, diff=%s", cmp.Diff(got, tc.expectedRenderedDp))
			}
		})
	}
}

// to make the image (which we change regularly) invariant for the tests
func fixSchedulerImage(dp *appsv1.Deployment) {
	cnt := &dp.Spec.Template.Spec.Containers[0] // shortcut
	cnt.Image = "test.com/image:latest"
}

var configTemplate string = `apiVersion: kubescheduler.config.k8s.io/v1beta2
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- schedulerName: topology-aware-scheduler
  plugins:
    filter:
      enabled:
      - name: NodeResourceTopologyMatch
    reserve:
      enabled:
       - name: NodeResourceTopologyMatch
    score:
      enabled:
        - name: NodeResourceTopologyMatch
  # optional plugin configs
  pluginConfig:
  - name: NodeResourceTopologyMatch
    args: {}`

const expectedSchedDeploymentDefault string = `---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    component: scheduler
  name: topology-aware-scheduler
  namespace: tas-scheduler
spec:
  replicas: 1
  selector:
    matchLabels:
      component: scheduler
  strategy: {}
  template:
    metadata:
      labels:
        component: scheduler
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: node-role.kubernetes.io/control-plane
                operator: Exists
      containers:
      - args:
        - /bin/kube-scheduler
        - --config=/etc/kubernetes/scheduler-config.yaml
        - --v=4
        image: test.com/image:latest
        imagePullPolicy: Always
        livenessProbe:
          httpGet:
            path: /healthz
            port: 10259
            scheme: HTTPS
          initialDelaySeconds: 15
        name: topology-aware-scheduler
        readinessProbe:
          httpGet:
            path: /healthz
            port: 10259
            scheme: HTTPS
        resources:
          limits:
            cpu: 200m
            memory: 500Mi
          requests:
            cpu: 200m
            memory: 500Mi
        volumeMounts:
        - mountPath: /etc/kubernetes
          name: scheduler-config
          readOnly: true
      serviceAccountName: topology-aware-scheduler
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/control-plane
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
      volumes:
      - configMap:
          name: scheduler-config
        name: scheduler-config
`

const expectedSchedDeploymentNonAffine string = `---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    component: scheduler
  name: topology-aware-scheduler
  namespace: tas-scheduler
spec:
  replicas: 1
  selector:
    matchLabels:
      component: scheduler
  strategy: {}
  template:
    metadata:
      labels:
        component: scheduler
    spec:
      containers:
      - args:
        - /bin/kube-scheduler
        - --config=/etc/kubernetes/scheduler-config.yaml
        - --v=4
        image: test.com/image:latest
        imagePullPolicy: Always
        livenessProbe:
          httpGet:
            path: /healthz
            port: 10259
            scheme: HTTPS
          initialDelaySeconds: 15
        name: topology-aware-scheduler
        readinessProbe:
          httpGet:
            path: /healthz
            port: 10259
            scheme: HTTPS
        resources:
          limits:
            cpu: 200m
            memory: 500Mi
          requests:
            cpu: 200m
            memory: 500Mi
        volumeMounts:
        - mountPath: /etc/kubernetes
          name: scheduler-config
          readOnly: true
      serviceAccountName: topology-aware-scheduler
      volumes:
      - configMap:
          name: scheduler-config
        name: scheduler-config
`

const expectedSchedDeploymentVerbose string = `---
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    component: scheduler
  name: topology-aware-scheduler
  namespace: tas-scheduler
spec:
  replicas: 1
  selector:
    matchLabels:
      component: scheduler
  strategy: {}
  template:
    metadata:
      labels:
        component: scheduler
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: node-role.kubernetes.io/control-plane
                operator: Exists
      containers:
      - args:
        - /bin/kube-scheduler
        - --config=/etc/kubernetes/scheduler-config.yaml
        - --v=6
        image: test.com/image:latest
        imagePullPolicy: Always
        livenessProbe:
          httpGet:
            path: /healthz
            port: 10259
            scheme: HTTPS
          initialDelaySeconds: 15
        name: topology-aware-scheduler
        readinessProbe:
          httpGet:
            path: /healthz
            port: 10259
            scheme: HTTPS
        resources:
          limits:
            cpu: 200m
            memory: 500Mi
          requests:
            cpu: 200m
            memory: 500Mi
        volumeMounts:
        - mountPath: /etc/kubernetes
          name: scheduler-config
          readOnly: true
      serviceAccountName: topology-aware-scheduler
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/control-plane
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
      volumes:
      - configMap:
          name: scheduler-config
        name: scheduler-config
`
