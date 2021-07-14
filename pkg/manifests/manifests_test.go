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
			obj, err := ServiceAccount(tc.component)
			if tc.expectError {
				if err == nil || obj != nil {
					t.Fatalf("nil err or non-nil obj=%v", obj)
				}
			}
		})
	}
}

func TestGetClusterRole(t *testing.T) {
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
			obj, err := ClusterRole(tc.component)
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

func TestGetSchedulerPluginConfigMap(t *testing.T) {
	obj, err := SchedulerPluginConfigMap()
	if obj == nil || err != nil {
		t.Fatalf("nil obj or non-nil err=%v", err)
	}
}

func TestGetSchedulerPluginDeployment(t *testing.T) {
	obj, err := SchedulerPluginDeployment()
	if obj == nil || err != nil {
		t.Fatalf("nil obj or non-nil err=%v", err)
	}
}

func TestGetSchedulerPluginClusterRoleBindingKubeScheduler(t *testing.T) {
	obj, err := SchedulerPluginClusterRoleBindingKubeScheduler()
	if obj == nil || err != nil {
		t.Fatalf("nil obj or non-nil err=%v", err)
	}
}

func TestGetSchedulerPluginClusterRoleBindingNodeResourceTopology(t *testing.T) {
	obj, err := SchedulerPluginClusterRoleBindingNodeResourceTopology()
	if obj == nil || err != nil {
		t.Fatalf("nil obj or non-nil err=%v", err)
	}
}

func TestGetSchedulerPluginClusterRoleBindingVolumeScheduler(t *testing.T) {
	obj, err := SchedulerPluginClusterRoleBindingVolumeScheduler()
	if obj == nil || err != nil {
		t.Fatalf("nil obj or non-nil err=%v", err)
	}
}

func TestGetSchedulerPluginRoleBindingKubeScheduler(t *testing.T) {
	obj, err := SchedulerPluginRoleBindingKubeScheduler()
	if obj == nil || err != nil {
		t.Fatalf("nil obj or non-nil err=%v", err)
	}
}

func TestGetResourceTopologyExporterClusterRoleBinding(t *testing.T) {
	obj, err := ResourceTopologyExporterClusterRoleBinding()
	if obj == nil || err != nil {
		t.Fatalf("nil obj or non-nil err=%v", err)
	}
}

func TestGetResourceTopologyExporterDaemonSet(t *testing.T) {
	obj, err := ResourceTopologyExporterDaemonSet()
	if obj == nil || err != nil {
		t.Fatalf("nil obj or non-nil err=%v", err)
	}
}
