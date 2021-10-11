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
)

func TestGetNamespace(t *testing.T) {
	type testCase struct {
		component   string
		expectError bool
	}

	testCases := []testCase{
		{
			component:   "unknown-wrong",
			expectError: true,
		},
		{
			component:   ComponentAPI,
			expectError: true,
		},
		{
			component:   ComponentSchedulerPlugin,
			expectError: false,
		},
		{
			component:   ComponentResourceTopologyExporter,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.component, func(t *testing.T) {
			obj, err := Namespace(tc.component)
			if tc.expectError {
				if err == nil || obj != nil {
					t.Fatalf("nil err or non-nil obj=%v", obj)
				}
			}
		})
	}
}

func TestGetServiceAccount(t *testing.T) {
	type testCase struct {
		component    string
		subComponent string
		expectError  bool
	}

	testCases := []testCase{
		{
			component:   "unknown-wrong",
			expectError: true,
		},
		{
			component:   ComponentAPI,
			expectError: true,
		},
		{
			component:    ComponentSchedulerPlugin,
			subComponent: SubComponentSchedulerPluginScheduler,
			expectError:  false,
		},
		{
			component:    ComponentSchedulerPlugin,
			subComponent: SubComponentSchedulerPluginController,
			expectError:  false,
		},
		{
			component:   ComponentResourceTopologyExporter,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.component, func(t *testing.T) {
			obj, err := ServiceAccount(tc.component, tc.subComponent)
			if tc.expectError {
				if err == nil || obj != nil {
					t.Fatalf("nil err or non-nil obj=%v", obj)
				}
			}
		})
	}
}

func TestGetRole(t *testing.T) {
	type testCase struct {
		component    string
		subComponent string
		expectError  bool
	}

	testCases := []testCase{
		{
			component:   "unknown-wrong",
			expectError: true,
		},
		{
			component:   ComponentAPI,
			expectError: true,
		},
		{
			component:    ComponentSchedulerPlugin,
			subComponent: SubComponentSchedulerPluginScheduler,
			expectError:  true,
		},
		{
			component:    ComponentSchedulerPlugin,
			subComponent: SubComponentSchedulerPluginController,
			expectError:  true,
		},
		{
			component:   ComponentResourceTopologyExporter,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.component, func(t *testing.T) {
			obj, err := Role(tc.component, tc.subComponent)
			if tc.expectError {
				if err == nil || obj != nil {
					t.Fatalf("nil err or non-nil obj=%v", obj)
				}
			} else {
				if err != nil || obj == nil {
					t.Fatalf("nil obj or non-nil err=%v", err)
				}
			}
		})
	}
}

func TestGetRoleBinding(t *testing.T) {
	type testCase struct {
		component    string
		subComponent string
		expectError  bool
	}

	testCases := []testCase{
		{
			component:   "unknown-wrong",
			expectError: true,
		},
		{
			component:   ComponentAPI,
			expectError: true,
		},
		{
			component:    ComponentSchedulerPlugin,
			subComponent: SubComponentSchedulerPluginScheduler,
			expectError:  true,
		},
		{
			component:    ComponentSchedulerPlugin,
			subComponent: SubComponentSchedulerPluginController,
			expectError:  true,
		},
		{
			component:   ComponentResourceTopologyExporter,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.component, func(t *testing.T) {
			obj, err := RoleBinding(tc.component, tc.subComponent)
			if tc.expectError {
				if err == nil || obj != nil {
					t.Fatalf("nil err or non-nil obj=%v", obj)
				}
			} else {
				if err != nil || obj == nil {
					t.Fatalf("nil obj or non-nil err=%v", err)
				}
			}
		})
	}
}

func TestGetClusterRole(t *testing.T) {
	type testCase struct {
		component    string
		subComponent string
		expectError  bool
	}

	testCases := []testCase{
		{
			component:   "unknown-wrong",
			expectError: true,
		},
		{
			component:   ComponentAPI,
			expectError: true,
		},
		{
			component:    ComponentSchedulerPlugin,
			subComponent: SubComponentSchedulerPluginScheduler,
			expectError:  false,
		},
		{
			component:    ComponentSchedulerPlugin,
			subComponent: SubComponentSchedulerPluginController,
			expectError:  false,
		},
		{
			component:   ComponentResourceTopologyExporter,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.component, func(t *testing.T) {
			obj, err := ClusterRole(tc.component, tc.subComponent)
			if tc.expectError {
				if err == nil || obj != nil {
					t.Fatalf("nil err or non-nil obj=%v", obj)
				}
			} else {
				if err != nil || obj == nil {
					t.Fatalf("nil obj or non-nil err=%v", err)
				}
			}
		})
	}
}

func TestGetClusterRoleBinding(t *testing.T) {
	type testCase struct {
		component    string
		subComponent string
		expectError  bool
	}

	testCases := []testCase{
		{
			component:   "unknown-wrong",
			expectError: true,
		},
		{
			component:   ComponentAPI,
			expectError: true,
		},
		{
			component:    ComponentSchedulerPlugin,
			subComponent: SubComponentSchedulerPluginScheduler,
			expectError:  false,
		},
		{
			component:    ComponentSchedulerPlugin,
			subComponent: SubComponentSchedulerPluginController,
			expectError:  false,
		},
		{
			component:   ComponentResourceTopologyExporter,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.component, func(t *testing.T) {
			obj, err := ClusterRoleBinding(tc.component, tc.subComponent)
			if tc.expectError {
				if err == nil || obj != nil {
					t.Fatalf("nil err or non-nil obj=%v", obj)
				}
			} else {
				if err != nil || obj == nil {
					t.Fatalf("nil obj or non-nil err=%v", err)
				}
			}
		})
	}
}

func TestGetAPICRD(t *testing.T) {
	obj, err := APICRD()
	if obj == nil || err != nil {
		t.Fatalf("nil obj or non-nil err=%v", err)
	}
}

func TestGetSchedulerCRD(t *testing.T) {
	obj, err := SchedulerCRD()
	if obj == nil || err != nil {
		t.Fatalf("nil obj or non-nil err=%v", err)
	}
}

func TestGetConfigMap(t *testing.T) {
	type testCase struct {
		component    string
		subComponent string
		expectError  bool
	}

	testCases := []testCase{
		{
			component:   "unknown-wrong",
			expectError: true,
		},
		{
			component:   ComponentAPI,
			expectError: true,
		},
		{
			component:   ComponentSchedulerPlugin,
			expectError: false,
		},
		{
			component:   ComponentResourceTopologyExporter,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.component, func(t *testing.T) {
			obj, err := ConfigMap(tc.component, tc.subComponent)
			if tc.expectError {
				if err == nil || obj != nil {
					t.Fatalf("nil err or non-nil obj=%v", obj)
				}
			} else {
				if err != nil || obj == nil {
					t.Fatalf("nil obj or non-nil err=%v", err)
				}
			}
		})
	}
}

func TestGetDeployment(t *testing.T) {
	type testCase struct {
		component    string
		subComponent string
		expectError  bool
	}

	testCases := []testCase{
		{
			component:   "unknown-wrong",
			expectError: true,
		},
		{
			component:   ComponentAPI,
			expectError: true,
		},
		{
			component:    ComponentSchedulerPlugin,
			subComponent: SubComponentSchedulerPluginScheduler,
			expectError:  false,
		},
		{
			component:    ComponentSchedulerPlugin,
			subComponent: SubComponentSchedulerPluginController,
			expectError:  false,
		},
		{
			component:   ComponentResourceTopologyExporter,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.component, func(t *testing.T) {
			obj, err := Deployment(tc.component, tc.subComponent)
			if tc.expectError {
				if err == nil || obj != nil {
					t.Fatalf("nil err or non-nil obj=%v", obj)
				}
			} else {
				if err != nil || obj == nil {
					t.Fatalf("nil obj or non-nil err=%v", err)
				}
			}
		})
	}
}

func TestGetDaemonSet(t *testing.T) {
	type testCase struct {
		component   string
		expectError bool
	}

	testCases := []testCase{
		{
			component:   "unknown-wrong",
			expectError: true,
		},
		{
			component:   ComponentAPI,
			expectError: true,
		},
		{
			component:   ComponentSchedulerPlugin,
			expectError: true,
		},
		{
			component:   ComponentResourceTopologyExporter,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.component, func(t *testing.T) {
			obj, err := DaemonSet(tc.component)
			if tc.expectError {
				if err == nil || obj != nil {
					t.Fatalf("nil err or non-nil obj=%v", obj)
				}
			} else {
				if err != nil || obj == nil {
					t.Fatalf("nil obj or non-nil err=%v", err)
				}
			}
		})
	}
}
