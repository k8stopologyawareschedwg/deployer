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
	"fmt"
	"strings"
	"testing"

	igntypes "github.com/coreos/ignition/v2/config/v3_2/types"
	"k8s.io/klog/v2"

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
			obj, err := RoleBinding(tc.component, tc.subComponent, "")
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
			_, err := DaemonSet(tc.component, tc.subComponent, platform.Kubernetes, "")
			if (err != nil) != tc.expectError {
				t.Fatalf("nil obj or non-nil err=%v", err)
			}
		})
	}
}

func TestDaemonSet(t *testing.T) {
	type testCase struct {
		name                 string
		plat                 platform.Platform
		expectedCommandArgs  []string
		expectedVolumes      map[string]string
		expectedVolumeMounts map[string]string
	}

	containerPodResourcesSocket := fmt.Sprintf("/%s/kubelet.sock", rtePodresourcesSocketVolumeName)
	containerHostSysDir := fmt.Sprintf("/%s", rteSysVolumeName)
	testCases := []testCase{
		{
			name: "Verify DaemonSet generation for OpenShift platform",
			plat: platform.OpenShift,
			expectedCommandArgs: []string{
				fmt.Sprintf("--sysfs=%s", containerHostSysDir),
				fmt.Sprintf("--podresources-socket=unix://%s", containerPodResourcesSocket),
				fmt.Sprintf("--notify-file=/%s/%s", rteNotifierVolumeName, rteNotifierFileName),
			},
			expectedVolumes: map[string]string{
				rteSysVolumeName:                "/sys",
				rtePodresourcesSocketVolumeName: "/var/lib/kubelet/pod-resources/kubelet.sock",
				rteNotifierVolumeName:           "/run/rte",
			},
			expectedVolumeMounts: map[string]string{
				rteSysVolumeName:                containerHostSysDir,
				rtePodresourcesSocketVolumeName: containerPodResourcesSocket,
				rteNotifierVolumeName:           fmt.Sprintf("/%s", rteNotifierVolumeName),
			},
		},
		{
			name: "Verify DaemonSet generation for Kubernetes platform",
			plat: platform.Kubernetes,
			expectedCommandArgs: []string{
				fmt.Sprintf("--sysfs=%s", containerHostSysDir),
				fmt.Sprintf("--podresources-socket=unix://%s", containerPodResourcesSocket),
				fmt.Sprintf("--kubelet-config-file=/%s/config.yaml", rteKubeletDirVolumeName),
				fmt.Sprintf("--kubelet-state-dir=/%s", rteKubeletDirVolumeName),
				fmt.Sprintf("--notify-file=/%s/%s", rteNotifierVolumeName, rteNotifierFileName),
			},
			expectedVolumes: map[string]string{
				rteSysVolumeName:                "/sys",
				rtePodresourcesSocketVolumeName: "/var/lib/kubelet/pod-resources/kubelet.sock",
				rteKubeletDirVolumeName:         "/var/lib/kubelet",
				rteNotifierVolumeName:           "/run/rte",
			},
			expectedVolumeMounts: map[string]string{
				rteSysVolumeName:                containerHostSysDir,
				rtePodresourcesSocketVolumeName: containerPodResourcesSocket,
				rteKubeletDirVolumeName:         fmt.Sprintf("/%s", rteKubeletDirVolumeName),
				rteNotifierVolumeName:           fmt.Sprintf("/%s", rteNotifierVolumeName),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ds, err := DaemonSet(ComponentResourceTopologyExporter, "", tc.plat, "test")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// we are expecting 3 volumes
			// 1. Host sys
			// 2. Pod resources socket file
			// 3. RTE notifier directory
			// 4. Kubelet directory only for Kubernetes platform
			expectedVolumesNumber := 3
			if tc.plat == platform.Kubernetes {
				expectedVolumesNumber = 4
			}
			if len(ds.Spec.Template.Spec.Volumes) != expectedVolumesNumber {
				klog.Errorf("the daemon set volumes: %+v", ds.Spec.Template.Spec.Volumes)
				t.Fatalf("the daemon set has %d volumes when it should have %d", len(ds.Spec.Template.Spec.Volumes), expectedVolumesNumber)
			}

			for _, v := range ds.Spec.Template.Spec.Volumes {
				path, ok := tc.expectedVolumes[v.Name]
				if !ok {
					t.Fatalf("the volume %q does not exist under expected volumes %v", v.Name, tc.expectedVolumes)
				}

				if v.HostPath.Path != path {
					t.Fatalf("the volume %q path %q does not have expected value %q", v.Name, v.HostPath.Path, path)
				}
			}

			rteContainer := ds.Spec.Template.Spec.Containers[0]
			if len(rteContainer.VolumeMounts) != expectedVolumesNumber {
				klog.Errorf("the daemon set container volume mounts: %+v", rteContainer.VolumeMounts)
				t.Fatalf("the daemon set container has %d volume mounts when it should have %d", len(rteContainer.VolumeMounts), expectedVolumesNumber)
			}

			for _, m := range rteContainer.VolumeMounts {
				path, ok := tc.expectedVolumeMounts[m.Name]
				if !ok {
					t.Fatalf("the volume mount %q does not exist under expected volumes mounts %v", m.Name, tc.expectedVolumeMounts)
				}

				if m.MountPath != path {
					t.Fatalf("the volume mount %q path %q does not have expected value %q", m.Name, m.MountPath, path)
				}
			}

			containerCommand := strings.Join(rteContainer.Args, " ")
			for _, arg := range tc.expectedCommandArgs {
				if !strings.Contains(containerCommand, arg) {
					t.Fatalf("the container command %q does not container argument %q", containerCommand, arg)
				}
			}
		})
	}
}

func TestMachineConfig(t *testing.T) {

	type testCase struct {
		name            string
		platformVersion platform.Version
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
			expectedFileNum: 3,
			expectedUnitNum: 1,
		},
		{
			name:            "OCP 4.11",
			platformVersion: "v4.11",
			expectedFileNum: 3,
			expectedUnitNum: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mcOpts := MachineConfigOptions{
				EnableNotifier: true,
			}
			mc, err := MachineConfig(ComponentResourceTopologyExporter, platform.Version(tc.platformVersion), mcOpts)
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
