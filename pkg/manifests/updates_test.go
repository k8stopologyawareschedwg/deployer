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
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
)

func TestUpdateResourceTopologyExporterCommandMultipleCalls(t *testing.T) {
	type testCase struct {
		name     string
		args     []string
		vars     map[string]string
		plat     platform.Platform
		expected []string
	}

	testCases := []testCase{
		{
			name:     "kubernetes, no vars",
			args:     []string{"/bin/k8sfoo", "-v=2", "--bar=42"},
			plat:     platform.Kubernetes,
			expected: []string{"/bin/k8sfoo", "-v=2", "--bar=42", "--kubelet-config-file=/host-var/lib/kubelet/config.yaml"},
		},
		{
			name:     "openshift, no vars",
			args:     []string{"/bin/ocpfoo", "-v=3", "--baz=42"},
			plat:     platform.OpenShift,
			expected: []string{"/bin/ocpfoo", "-v=3", "--baz=42", "--topology-manager-policy=single-numa-node"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// the most important magic numbers are: 0, 1, 2, ... N
			// we avoid 0 for obvious reasons, we limit N=10 for practicalty.
			// Hence: 1, 2, N=10
			iterations := []int{1, 2, 10}
			for _, its := range iterations {
				args := append([]string{}, tc.args...)
				var retArgs []string
				for idx := 0; idx < its; idx++ {
					retArgs = UpdateResourceTopologyExporterCommand(args, tc.vars, tc.plat)
					args = retArgs
				}

				if !reflect.DeepEqual(tc.expected, retArgs) {
					t.Errorf("testcase %q iterations %d expected %v got %v", tc.name, its, tc.expected, retArgs)
				}
			}
		})
	}
}

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
