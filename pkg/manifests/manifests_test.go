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
	"encoding/json"
	"testing"

	igntypes "github.com/coreos/ignition/v2/config/v3_2/types"
	"k8s.io/klog/v2"

	selinuxassets "github.com/k8stopologyawareschedwg/deployer/pkg/assets/selinux"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
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
			obj, err := ServiceAccount(tc.component, tc.subComponent, "")
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
			expectError:  false,
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
			obj, err := Role(tc.component, tc.subComponent, "")
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
		roleName     string
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
			subComponent: SubComponentSchedulerPluginController,
			roleName:     "",
			expectError:  false,
		},
		{
			component:    ComponentSchedulerPlugin,
			subComponent: SubComponentSchedulerPluginScheduler,
			roleName:     "",
			expectError:  true,
		},
		{
			component:    ComponentSchedulerPlugin,
			subComponent: SubComponentSchedulerPluginScheduler,
			roleName:     RoleNameAuthReader,
			expectError:  false,
		},
		{
			component:    ComponentSchedulerPlugin,
			subComponent: SubComponentSchedulerPluginScheduler,
			roleName:     RoleNameLeaderElect,
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
			obj, err := RoleBinding(tc.component, tc.subComponent, tc.roleName, "")
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
			expectError: false,
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
			expectError: false,
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
			obj, err := Deployment(tc.component, tc.subComponent, "")
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
			expectError: true,
		},
		{
			component:   ComponentResourceTopologyExporter,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.component, func(t *testing.T) {
			_, err := DaemonSet(tc.component, tc.subComponent, "")
			if (err != nil) != tc.expectError {
				t.Fatalf("nil obj or non-nil err=%v", err)
			}
		})
	}
}

func TestMachineConfig(t *testing.T) {

	type testCase struct {
		name            string
		platformVersion platform.Version
		enableCRIHooks  bool
		expectedFileNum int
		expectedUnitNum int
	}
	// In both these cases:
	// we are expecting to have 3 files
	// 1. OCI hook configuration
	// 2. OCI hook script
	// 3. SELinux policy

	// One systemd unit
	// 1. Systemd unit to install the SELinux policy

	// TODO: Check SELinuxPolicy in the various cases
	testCases := []testCase{
		{
			name:            "OCP 4.10",
			platformVersion: "v4.10",
			enableCRIHooks:  true,
			expectedFileNum: 3,
			expectedUnitNum: 1,
		},
		{
			name:            "OCP 4.11",
			platformVersion: "v4.11",
			enableCRIHooks:  true,
			expectedFileNum: 3,
			expectedUnitNum: 1,
		},
		{
			name:            "OCP 4.11",
			platformVersion: "v4.11",
			enableCRIHooks:  false,
			expectedFileNum: 1,
			expectedUnitNum: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mc, err := MachineConfig(ComponentResourceTopologyExporter, platform.Version(tc.platformVersion), tc.enableCRIHooks)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			ignitionConfig := &igntypes.Config{}
			if err := json.Unmarshal(mc.Spec.Config.Raw, ignitionConfig); err != nil {
				t.Fatalf("failed to unmarshal ignition config: %v", err)
			}

			if len(ignitionConfig.Storage.Files) != tc.expectedFileNum {
				klog.Errorf("ignition config files: %+v", ignitionConfig.Storage.Files)
				t.Fatalf("the ignition config has %d files when it should have %d", len(ignitionConfig.Storage.Files), tc.expectedFileNum)
			}

			if len(ignitionConfig.Systemd.Units) != tc.expectedUnitNum {
				klog.Errorf("ignition config systemd units: %+v", ignitionConfig.Systemd.Units)
				t.Fatalf("the ignition config has %d systemd units when it should have %d", len(ignitionConfig.Systemd.Units), tc.expectedUnitNum)
			}
		})
	}
}

func TestSecurityContextConstraint(t *testing.T) {
	testCases := []struct {
		description             string
		withCustomSELinuxPolicy bool
		selinuxContextType      string
	}{
		{
			description:             "with custom (legacy) policy",
			withCustomSELinuxPolicy: true,
			selinuxContextType:      selinuxassets.RTEContextTypeLegacy,
		},
		{
			description:             "with built-in policy",
			withCustomSELinuxPolicy: false,
			selinuxContextType:      selinuxassets.RTEContextType,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			scc, err := SecurityContextConstraint(ComponentResourceTopologyExporter, tc.withCustomSELinuxPolicy)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if scc == nil {
				t.Fatalf("nil security context constraint")
			}
			// shortcut
			scType := scc.SELinuxContext.SELinuxOptions.Type
			if scType != tc.selinuxContextType {
				t.Fatalf("wrong selinux context type; got=%s want=%s", scType, tc.selinuxContextType)
			}
		})
	}
}

func TestNetworkPolicy(t *testing.T) {
	testCases := []struct {
		description  string
		policyType   string
		policyName   string
		component    string
		subComponent string
		policyExist  bool
	}{
		{
			description:  "rte default network policy",
			policyType:   "default",
			policyName:   "rte-default-deny-all",
			component:    ComponentResourceTopologyExporter,
			subComponent: "",
			policyExist:  true,
		},
		{
			description:  "rte api server network policy",
			policyType:   "apiserver",
			policyName:   "rte-egress-to-api-server",
			component:    ComponentResourceTopologyExporter,
			subComponent: "",
			policyExist:  true,
		},
		{
			description:  "rte metrics server network policy",
			policyType:   "metrics",
			policyName:   "ingress-to-rte-metrics",
			component:    ComponentResourceTopologyExporter,
			subComponent: "",
			policyExist:  true,
		},
		{
			description:  "scheduler controller default network policy",
			policyType:   "default",
			policyName:   "topology-aware-controller-default-deny-all",
			component:    ComponentSchedulerPlugin,
			subComponent: SubComponentSchedulerPluginController,
			policyExist:  true,
		},
		{
			description:  "topology-aware-controller api server network policy",
			policyType:   "apiserver",
			policyName:   "topology-aware-controller-egress-to-api-server",
			component:    ComponentSchedulerPlugin,
			subComponent: SubComponentSchedulerPluginController,
			policyExist:  true,
		},
		{
			description:  "scheduler default network policy",
			policyType:   "default",
			policyName:   "topology-aware-scheduler-default-deny-all",
			component:    ComponentSchedulerPlugin,
			subComponent: SubComponentSchedulerPluginScheduler,
			policyExist:  true,
		},
		{
			description:  "topology-aware-scheduler api server network policy",
			policyType:   "apiserver",
			policyName:   "topology-aware-scheduler-egress-to-api-server",
			component:    ComponentSchedulerPlugin,
			subComponent: SubComponentSchedulerPluginScheduler,
			policyExist:  true,
		},
		{
			description:  "wrong network policy",
			policyType:   "Non existent policy",
			policyName:   "Non existent policy",
			component:    "",
			subComponent: "",
			policyExist:  false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			np, err := NetworkPolicy(tc.component, tc.subComponent, tc.policyType, "numaresources-operator")
			if !tc.policyExist {
				if err == nil {
					t.Fatalf("expected error for non-existent policy %q, but got none", tc.policyType)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.policyType, err)
			}

			if np == nil {
				t.Fatal("expected a network policy but got nil")
			}

			if np.Name != tc.policyName {
				t.Fatalf("unexpected network policy name: got=%q, want=%q", np.Name, tc.policyName)
			}
		})
	}
}
