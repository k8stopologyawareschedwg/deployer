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
