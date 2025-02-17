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

package objectupdate

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func TestFindContainerByName(t *testing.T) {

	type testCase struct {
		name          string
		containers    []corev1.Container
		targetName    string
		expectedFound bool
	}

	testCases := []testCase{
		{
			name:       "nil list",
			targetName: "foo",
		},
		{
			name:       "empty list",
			containers: []corev1.Container{},
			targetName: "foo",
		},
		{
			name: "not found",
			containers: []corev1.Container{
				{
					Name: "bar",
				},
				{
					Name: "baz",
				},
			},
			targetName: "foo",
		},
		{
			name: "found",
			containers: []corev1.Container{
				{
					Name: "foo",
				},
				{
					Name: "bar",
				},
				{
					Name: "baz",
				},
			},
			targetName:    "bar",
			expectedFound: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := FindContainerByName(tc.containers, tc.targetName)
			found := (got != nil)
			if found != tc.expectedFound {
				t.Errorf("container found=%v expected=%v", found, tc.expectedFound)
			}
		})
	}
}

func TestFindContainerByNameMutablePod(t *testing.T) {
	testImageName := "test.io/image"

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "foo",
				},
				{
					Name: "bar",
				},
				{
					Name: "baz",
				},
			},
		},
	}

	got := FindContainerByName(pod.Spec.Containers, "bar")
	if got == nil {
		t.Fatalf("missing container")
	}

	got.Image = testImageName

	// intentionally hardcode the path
	if pod.Spec.Containers[1].Image != testImageName {
		t.Fatalf("failed to mutate through the FindContainerByName reference")
	}
}

func TestFindContainerByNameMutableDeployment(t *testing.T) {
	testWorkingDir := "/foo/bar"

	dp := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "foo",
						},
						{
							Name: "bar",
						},
					},
				},
			},
		},
	}

	got := FindContainerByName(dp.Spec.Template.Spec.Containers, "foo")
	if got == nil {
		t.Fatalf("missing container")
	}

	got.WorkingDir = testWorkingDir

	// intentionally hardcode the path
	if dp.Spec.Template.Spec.Containers[0].WorkingDir != testWorkingDir {
		t.Fatalf("failed to mutate through the FindContainerByName reference")
	}
}

func TestFindContainerEnvVarByNameMutablePod(t *testing.T) {
	testEnvVarName := "TEST_FOO_BAR"

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "foo",
					Env: []corev1.EnvVar{
						{
							Name:  testEnvVarName,
							Value: "33",
						},
					},
				},
			},
		},
	}

	got := FindContainerEnvVarByName(pod.Spec.Containers[0].Env, testEnvVarName)
	if got == nil {
		t.Fatalf("missing container env var")
	}

	newValue := "42"
	got.Value = newValue
	if pod.Spec.Containers[0].Env[0].Value != newValue {
		t.Fatalf("failed to mutate through the FindContainerEnvVarByName reference")
	}
}

func TestFindContainerEnvVarByNameMutableDeployment(t *testing.T) {
	testEnvVarName := "FIZZBUZZ"

	dp := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "foo",
							Env: []corev1.EnvVar{
								{
									Name:  testEnvVarName,
									Value: "27",
								},
							},
						},
					},
				},
			},
		},
	}

	got := FindContainerEnvVarByName(dp.Spec.Template.Spec.Containers[0].Env, testEnvVarName)
	if got == nil {
		t.Fatalf("missing container env var")
	}

	newValue := "42"
	got.Value = newValue
	if dp.Spec.Template.Spec.Containers[0].Env[0].Value != newValue {
		t.Fatalf("failed to mutate through the FindContainerEnvVarByName reference")
	}
}

func TestFindContainerPortByNameMutablePod(t *testing.T) {
	testPortName := "foo-port"

	pod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "foo",
					Ports: []corev1.ContainerPort{
						{
							Name:          testPortName,
							ContainerPort: int32(12345),
						},
					},
				},
			},
		},
	}

	got := FindContainerPortByName(pod.Spec.Containers[0].Ports, testPortName)
	if got == nil {
		t.Fatalf("missing container env var")
	}

	newValue := int32(42224)
	got.ContainerPort = newValue
	if pod.Spec.Containers[0].Ports[0].ContainerPort != newValue {
		t.Fatalf("failed to mutate through the FindContainerPortByName reference")
	}
}

func TestFindContainerPortByNameMutableDeployment(t *testing.T) {
	testPortName := "bar-port"

	dp := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "foo",

							Ports: []corev1.ContainerPort{
								{
									Name:          testPortName,
									ContainerPort: int32(12345),
								},
							},
						},
					},
				},
			},
		},
	}

	got := FindContainerPortByName(dp.Spec.Template.Spec.Containers[0].Ports, testPortName)
	if got == nil {
		t.Fatalf("missing container env var")
	}

	newValue := int32(24242)
	got.ContainerPort = newValue
	if dp.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort != newValue {
		t.Fatalf("failed to mutate through the FindContainerPortByName reference")
	}
}
