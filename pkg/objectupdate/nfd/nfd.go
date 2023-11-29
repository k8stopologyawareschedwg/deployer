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
	"fmt"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/k8stopologyawareschedwg/deployer/pkg/flagcodec"
	"github.com/k8stopologyawareschedwg/deployer/pkg/images"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
	"github.com/k8stopologyawareschedwg/deployer/pkg/objectupdate"
)

func UpdaterDaemonSet(ds *appsv1.DaemonSet, opts objectupdate.DaemonSetOptions) {
	if c := objectupdate.FindContainerByName(ds.Spec.Template.Spec.Containers, manifests.ContainerNameNFDTopologyUpdater); c != nil {
		c.ImagePullPolicy = corev1.PullAlways
		if opts.PullIfNotPresent {
			c.ImagePullPolicy = corev1.PullIfNotPresent
		}

		flags := flagcodec.ParseArgvKeyValue(c.Args, flagcodec.WithFlagNormalization)
		flags.SetOption("-v", fmt.Sprintf("%d", opts.Verbose))
		if opts.UpdateInterval > 0 {
			flags.SetOption("--sleep-interval", fmt.Sprintf("%v", opts.UpdateInterval))
		} else {
			flags.Delete("--sleep-interval")
		}

		flags.SetOption("--pods-fingerprint", strconv.FormatBool(opts.PFPEnable))

		// we need to explicitly disable the kubelet state dir monitoring, which is opt-out
		flags.SetOption("--kubelet-state-dir", "")

		c.Args = flags.Argv()

		c.Image = images.NodeFeatureDiscoveryImage
	}

	if opts.NodeSelector != nil {
		ds.Spec.Template.Spec.NodeSelector = opts.NodeSelector.MatchLabels
	}
}
