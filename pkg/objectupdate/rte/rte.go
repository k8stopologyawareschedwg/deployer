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
	"strconv"

	"github.com/k8stopologyawareschedwg/deployer/pkg/flagcodec"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
	"github.com/k8stopologyawareschedwg/deployer/pkg/objectupdate"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	metricsPort        = 2112
	rteConfigMountName = "rte-config-volume"
	RTEConfigMapName   = "rte-config"
)

func ContainerConfig(podSpec *corev1.PodSpec, cnt *corev1.Container, configMapName string) {
	cnt.VolumeMounts = append(cnt.VolumeMounts,
		corev1.VolumeMount{
			Name:      rteConfigMountName,
			MountPath: "/etc/resource-topology-exporter/",
		},
	)
	podSpec.Volumes = append(podSpec.Volumes,
		corev1.Volume{
			Name: rteConfigMountName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: configMapName,
					},
					Optional: newBool(true),
				},
			},
		},
	)
}

func DaemonSet(ds *appsv1.DaemonSet, configMapName string, opts objectupdate.DaemonSetOptions) {
	podSpec := &ds.Spec.Template.Spec
	if cntSpec := objectupdate.FindContainerByName(ds.Spec.Template.Spec.Containers, manifests.ContainerNameRTE); cntSpec != nil {

		cntSpec.ImagePullPolicy = corev1.PullAlways
		if opts.PullIfNotPresent {
			cntSpec.ImagePullPolicy = corev1.PullIfNotPresent
		}

		flags := flagcodec.ParseArgvKeyValue(cntSpec.Args)
		if opts.PFPEnable {
			flags.SetToggle("--pods-fingerprint")
		}
		cntSpec.Args = flags.Argv()

		if configMapName != "" {
			ContainerConfig(podSpec, cntSpec, configMapName)
		}
	}

	if opts.NodeSelector != nil {
		podSpec.NodeSelector = opts.NodeSelector.MatchLabels
	}
	MetricsPort(ds, metricsPort)
}

func MetricsPort(ds *appsv1.DaemonSet, pNum int) {
	cntSpec := objectupdate.FindContainerByName(ds.Spec.Template.Spec.Containers, manifests.ContainerNameRTE)
	if cntSpec == nil {
		return
	}

	pNumAsStr := strconv.Itoa(pNum)

	for idx, env := range cntSpec.Env {
		if env.Name == "METRICS_PORT" {
			cntSpec.Env[idx].Value = pNumAsStr
		}
	}

	cp := []corev1.ContainerPort{{
		Name:          "metrics-port",
		ContainerPort: int32(pNum),
	},
	}
	cntSpec.Ports = cp
}

func newBool(val bool) *bool {
	return &val
}
