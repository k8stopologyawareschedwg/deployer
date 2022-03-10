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
 * Copyright 2021 Red Hat, Inc.
 */

package manifests

import (
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func TestUpdateMetricsPort(t *testing.T) {
	ds := &appsv1.DaemonSet{
		Spec: appsv1.DaemonSetSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Env: []v1.EnvVar{
								{
									Name:  "METRIC_PORTS",
									Value: "${METRIC_PORTS}",
								},
							},
							Ports: []v1.ContainerPort{
								{
									Name: "metrics-port",
									// Must be a number so let's put something arbitrary
									ContainerPort: int32(1),
								},
							},
						},
					},
				},
			},
		},
	}

	type testCase struct {
		port  int
		sPort string
	}

	testCases := []testCase{
		{
			port:  3333,
			sPort: "3333",
		},
		{
			port:  1234,
			sPort: "1234",
		},
		{
			port:  2112,
			sPort: "2112",
		},
	}

	for _, tc := range testCases {
		t.Run("update metrics", func(t *testing.T) {
			UpdateMetricsPort(ds, tc.port)
			for _, env := range ds.Spec.Template.Spec.Containers[0].Env {
				if env.Name == "METRICS_PORT" && env.Value != tc.sPort {
					t.Errorf("expected port number to be %q got %q", tc.sPort, env.Value)
				}
			}

			for _, port := range ds.Spec.Template.Spec.Containers[0].Ports {
				if port.Name == "metrics-port" && port.ContainerPort != int32(tc.port) {
					t.Errorf("expected port number to be %d got %d", tc.port, port.ContainerPort)
				}
			}
		})
	}
}

func TestAddConfigMapToDaemonSet(t *testing.T) {
	ds := appsv1.DaemonSet{
		Spec: appsv1.DaemonSetSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{},
					},
				},
			},
		},
	}
	if len(ds.Spec.Template.Spec.Containers[0].VolumeMounts) != 0 {
		t.Errorf("unexpected volume mount")
	}
	if len(ds.Spec.Template.Spec.Volumes) != 0 {
		t.Errorf("unexpected volume declaration")
	}

	UpdateResourceTopologyExporterContainerConfig(&ds.Spec.Template.Spec, &ds.Spec.Template.Spec.Containers[0], "test-cfg")
	if len(ds.Spec.Template.Spec.Containers[0].VolumeMounts) != 1 {
		t.Errorf("missing volume mount")
	}
	if len(ds.Spec.Template.Spec.Volumes) != 1 {
		t.Errorf("missing volume declaration")
	}
}

func TestAddConfigMapToPod(t *testing.T) {
	pod := &v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{},
			},
		},
	}
	if len(pod.Spec.Containers[0].VolumeMounts) != 0 {
		t.Errorf("unexpected volume mount")
	}
	if len(pod.Spec.Volumes) != 0 {
		t.Errorf("unexpected volume declaration")
	}

	UpdateResourceTopologyExporterContainerConfig(&pod.Spec, &pod.Spec.Containers[0], "test-cfg")
	if len(pod.Spec.Containers[0].VolumeMounts) != 1 {
		t.Errorf("missing volume mount")
	}
	if len(pod.Spec.Volumes) != 1 {
		t.Errorf("missing volume declaration")
	}
}

func TestUpdateNFDTopologyUpdaterDaemonSet(t *testing.T) {
	ds := &appsv1.DaemonSet{
		Spec: appsv1.DaemonSetSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{},
					},
				},
			},
		},
	}

	testCases := []struct {
		cntName          string
		pullIfNotPresent bool
		nodeSelector     *metav1.LabelSelector
	}{
		{
			cntName:          containerNameNFDTopologyUpdater,
			pullIfNotPresent: false,
			nodeSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"foo": "bar"},
			},
		},
		{
			cntName:          containerNameNFDMaster,
			pullIfNotPresent: true,
			nodeSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"foo": "bar"},
			},
		},
	}
	for _, tc := range testCases {
		mutatedDs := ds.DeepCopy()
		pSpec := &mutatedDs.Spec.Template.Spec
		pSpec.Containers[0].Name = tc.cntName
		UpdateNFDTopologyUpdaterDaemonSet(mutatedDs, tc.pullIfNotPresent, tc.nodeSelector)
		if tc.cntName == containerNameNFDTopologyUpdater {
			if pSpec.Containers[0].ImagePullPolicy != pullPolicy(tc.pullIfNotPresent) {
				t.Errorf("expected container ImagePullPolicy to be: %q; got: %q", pullPolicy(tc.pullIfNotPresent), pSpec.Containers[0].ImagePullPolicy)
			}
			if !cmp.Equal(pSpec.NodeSelector, tc.nodeSelector.MatchLabels) {
				t.Errorf("expected NodeSelector to be: %v; got: %v", tc.nodeSelector.MatchLabels, pSpec.NodeSelector)
			}
		} else {
			if pSpec.Containers[0].ImagePullPolicy != "" {
				t.Errorf("container name is other than %q, no changes to container are expected", containerNameNFDTopologyUpdater)
			}
		}
	}
}

func TestUpdateNFDMasterDeployment(t *testing.T) {
	dp := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{},
					},
				},
			},
		},
	}

	testCases := []struct {
		cntName          string
		pullIfNotPresent bool
	}{
		{
			cntName:          containerNameNFDMaster,
			pullIfNotPresent: true,
		},
		{
			cntName:          "foo",
			pullIfNotPresent: true,
		},
	}
	for _, tc := range testCases {
		mutatedDp := dp.DeepCopy()
		pSpec := &mutatedDp.Spec.Template.Spec
		pSpec.Containers[0].Name = tc.cntName
		UpdateNFDMasterDeployment(mutatedDp, tc.pullIfNotPresent)
		if tc.cntName == containerNameNFDMaster {
			if pSpec.Containers[0].ImagePullPolicy != pullPolicy(tc.pullIfNotPresent) {
				t.Errorf("expected container ImagePullPolicy to be: %q; got: %q", pullPolicy(tc.pullIfNotPresent), pSpec.Containers[0].ImagePullPolicy)
			}
		} else {
			if pSpec.Containers[0].ImagePullPolicy != "" {
				t.Errorf("container name is other than %q, no changes to container are expected", containerNameNFDMaster)
			}
		}
	}
}
