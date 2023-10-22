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

	"github.com/google/go-cmp/cmp"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
)

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
			expectedRenderedDp: expectedSchedDeploymentNonAffine,
		},
	}

	dpRef, err := manifests.Deployment(manifests.ComponentSchedulerPlugin, manifests.SubComponentSchedulerPluginScheduler, "")
	if err != nil {
		t.Errorf("cannot load the scheduler manifest: %v", err)
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dp := dpRef.DeepCopy()

			SchedulerDeployment(dp, tc.pullIfNotPresent)
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
