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

package rte

import (
	"fmt"
	selinuxassets "github.com/k8stopologyawareschedwg/deployer/pkg/assets/selinux"
	"github.com/k8stopologyawareschedwg/deployer/pkg/objectupdate"
	"strings"
	"testing"
	"time"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
	"github.com/k8stopologyawareschedwg/deployer/pkg/options"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

func TestMetricsPort(t *testing.T) {
	ds := &appsv1.DaemonSet{
		Spec: appsv1.DaemonSetSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name: "resource-topology-exporter",
							Env: []v1.EnvVar{
								{
									Name:  "METRIC_PORTS",
									Value: "9999",
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
			MetricsPort(ds, tc.port)
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

	ContainerConfig(&ds.Spec.Template.Spec, &ds.Spec.Template.Spec.Containers[0], "test-cfg")
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

	ContainerConfig(&pod.Spec, &pod.Spec.Containers[0], "test-cfg")
	if len(pod.Spec.Containers[0].VolumeMounts) != 1 {
		t.Errorf("missing volume mount")
	}
	if len(pod.Spec.Volumes) != 1 {
		t.Errorf("missing volume declaration")
	}
}

func TestDaemonSet(t *testing.T) {
	type testCase struct {
		name                 string
		plat                 platform.Platform
		pfpEnable            bool
		expectedCommandArgs  []string
		expectedVolumes      map[string]string
		expectedVolumeMounts map[string]string
	}

	containerHostSysDir := fmt.Sprintf("/%s", rteSysVolumeName)
	testCases := []testCase{
		{
			name:      "Verify DaemonSet generation for OpenShift platform",
			plat:      platform.OpenShift,
			pfpEnable: true,
			expectedCommandArgs: []string{
				fmt.Sprintf("--sysfs=%s", containerHostSysDir),
				fmt.Sprintf("--podresources-socket=unix:///%s/%s", rtePodresourcesDirVolumeName, "kubelet.sock"),
				fmt.Sprintf("--notify-file=/%s/%s", rteNotifierVolumeName, rteNotifierFileName),
				"--pods-fingerprint=true",
			},
			expectedVolumes: map[string]string{
				rteSysVolumeName:             "/sys",
				rtePodresourcesDirVolumeName: "/var/lib/kubelet/pod-resources",
				rteNotifierVolumeName:        "/run/rte",
			},
			expectedVolumeMounts: map[string]string{
				rteSysVolumeName:             containerHostSysDir,
				rtePodresourcesDirVolumeName: fmt.Sprintf("/%s", rtePodresourcesDirVolumeName),
				rteNotifierVolumeName:        fmt.Sprintf("/%s", rteNotifierVolumeName),
			},
		},
		{
			name: "Verify DaemonSet generation for OpenShift platform with PFP disabled",
			plat: platform.OpenShift,
			expectedCommandArgs: []string{
				fmt.Sprintf("--sysfs=%s", containerHostSysDir),
				fmt.Sprintf("--podresources-socket=unix:///%s/%s", rtePodresourcesDirVolumeName, "kubelet.sock"),
				fmt.Sprintf("--notify-file=/%s/%s", rteNotifierVolumeName, rteNotifierFileName),
				"--pods-fingerprint=false",
			},
			expectedVolumes: map[string]string{
				rteSysVolumeName:             "/sys",
				rtePodresourcesDirVolumeName: "/var/lib/kubelet/pod-resources",
				rteNotifierVolumeName:        "/run/rte",
			},
			expectedVolumeMounts: map[string]string{
				rteSysVolumeName:             containerHostSysDir,
				rtePodresourcesDirVolumeName: fmt.Sprintf("/%s", rtePodresourcesDirVolumeName),
				rteNotifierVolumeName:        fmt.Sprintf("/%s", rteNotifierVolumeName),
			},
		},
		{
			name:      "Verify DaemonSet generation for Kubernetes platform",
			plat:      platform.Kubernetes,
			pfpEnable: true,
			expectedCommandArgs: []string{
				fmt.Sprintf("--sysfs=%s", containerHostSysDir),
				fmt.Sprintf("--podresources-socket=unix:///%s/%s", rtePodresourcesDirVolumeName, "kubelet.sock"),
				fmt.Sprintf("--kubelet-config-file=/%s/config.yaml", rteKubeletDirVolumeName),
				fmt.Sprintf("--notify-file=/%s/%s", rteNotifierVolumeName, rteNotifierFileName),
				"--pods-fingerprint=true",
			},
			expectedVolumes: map[string]string{
				rteSysVolumeName:             "/sys",
				rtePodresourcesDirVolumeName: "/var/lib/kubelet/pod-resources",
				rteKubeletDirVolumeName:      "/var/lib/kubelet",
				rteNotifierVolumeName:        "/run/rte",
			},
			expectedVolumeMounts: map[string]string{
				rteSysVolumeName:             containerHostSysDir,
				rtePodresourcesDirVolumeName: fmt.Sprintf("/%s", rtePodresourcesDirVolumeName),
				rteKubeletDirVolumeName:      fmt.Sprintf("/%s", rteKubeletDirVolumeName),
				rteNotifierVolumeName:        fmt.Sprintf("/%s", rteNotifierVolumeName),
			},
		},
		{
			name: "Verify DaemonSet generation for Kubernetes platform with PFP disabled",
			plat: platform.Kubernetes,
			expectedCommandArgs: []string{
				fmt.Sprintf("--sysfs=%s", containerHostSysDir),
				fmt.Sprintf("--podresources-socket=unix:///%s/%s", rtePodresourcesDirVolumeName, "kubelet.sock"),
				fmt.Sprintf("--kubelet-config-file=/%s/config.yaml", rteKubeletDirVolumeName),
				fmt.Sprintf("--notify-file=/%s/%s", rteNotifierVolumeName, rteNotifierFileName),
				"--pods-fingerprint=false",
			},
			expectedVolumes: map[string]string{
				rteSysVolumeName:             "/sys",
				rtePodresourcesDirVolumeName: "/var/lib/kubelet/pod-resources",
				rteKubeletDirVolumeName:      "/var/lib/kubelet",
				rteNotifierVolumeName:        "/run/rte",
			},
			expectedVolumeMounts: map[string]string{
				rteSysVolumeName:             containerHostSysDir,
				rtePodresourcesDirVolumeName: fmt.Sprintf("/%s", rtePodresourcesDirVolumeName),
				rteKubeletDirVolumeName:      fmt.Sprintf("/%s", rteKubeletDirVolumeName),
				rteNotifierVolumeName:        fmt.Sprintf("/%s", rteNotifierVolumeName),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ds, err := manifests.DaemonSet(manifests.ComponentResourceTopologyExporter, "", "test")
			if err != nil {
				t.Fatalf("unexpected error getting the manifests: %v", err)
			}
			DaemonSet(ds, tc.plat, "", options.DaemonSet{
				PFPEnable:          tc.pfpEnable,
				NotificationEnable: true,
				UpdateInterval:     10 * time.Second,
			})

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

func TestSecurityContext(t *testing.T) {
	testCases := []struct {
		description        string
		selinuxContextType string
	}{
		{
			description:        "custom policy",
			selinuxContextType: selinuxassets.RTEContextTypeLegacy,
		},
		{
			description:        "built-in policy",
			selinuxContextType: selinuxassets.RTEContextType,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			ds, err := manifests.DaemonSet(manifests.ComponentResourceTopologyExporter, "", "test")
			if err != nil {
				t.Fatalf("unexpected error getting the manifests: %v", err)
			}
			DaemonSet(ds, platform.OpenShift, "", options.DaemonSet{})
			SecurityContext(ds, tc.selinuxContextType)
			cntSpec := objectupdate.FindContainerByName(ds.Spec.Template.Spec.Containers, manifests.ContainerNameRTE)
			sc := cntSpec.SecurityContext
			if sc == nil {
				t.Fatalf("the security context for container %q does not exist", cntSpec.Name)
			}
			if sc.SELinuxOptions.Type != tc.selinuxContextType {
				t.Fatalf("wrong security context for container %q; want=%s got=%s", cntSpec.Name, tc.selinuxContextType, sc.SELinuxOptions.Type)
			}
		})
	}
}
