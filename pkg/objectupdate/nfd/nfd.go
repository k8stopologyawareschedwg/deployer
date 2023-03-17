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

package nfd

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/k8stopologyawareschedwg/deployer/pkg/images"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
)

func UpdaterDaemonSet(ds *appsv1.DaemonSet, pullIfNotPresent, pfpEnable bool, nodeSelector *metav1.LabelSelector) {
	for i := range ds.Spec.Template.Spec.Containers {
		c := &ds.Spec.Template.Spec.Containers[i]
		if c.Name != manifests.ContainerNameNFDTopologyUpdater {
			continue
		}
		c.ImagePullPolicy = corev1.PullAlways
		if pullIfNotPresent {
			c.ImagePullPolicy = corev1.PullIfNotPresent
		}

		if pfpEnable {
			c.Args = append([]string{"--pods-fingerprint"}, c.Args...)
		}

		c.Image = images.NodeFeatureDiscoveryImage

	}
	if nodeSelector != nil {
		ds.Spec.Template.Spec.NodeSelector = nodeSelector.MatchLabels
	}
}
