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
