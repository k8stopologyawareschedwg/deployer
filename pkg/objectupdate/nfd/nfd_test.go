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
	"testing"

	"github.com/google/go-cmp/cmp"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
	"github.com/k8stopologyawareschedwg/deployer/pkg/options"
)

func TestUpdaterDaemonSet(t *testing.T) {
	ds := &appsv1.DaemonSet{
		Spec: appsv1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{},
					},
				},
			},
		},
	}

	testCases := []struct {
		cntName             string
		pfpEnable           bool
		pullIfNotPresent    bool
		nodeSelector        *metav1.LabelSelector
		expectedCommandArgs []string
	}{
		{
			cntName:          manifests.ContainerNameNFDTopologyUpdater,
			pfpEnable:        true,
			pullIfNotPresent: false,
			nodeSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"foo": "bar"},
			},
			expectedCommandArgs: []string{
				"--kubelet-state-dir=",
				"--pods-fingerprint=true",
			},
		},
		{
			cntName:          manifests.ContainerNameNFDTopologyUpdater,
			pfpEnable:        false,
			pullIfNotPresent: false,
			nodeSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"foo": "bar"},
			},
			expectedCommandArgs: []string{
				"--kubelet-state-dir=",
				"--pods-fingerprint=false",
			},
		},
	}
	for _, tc := range testCases {
		mutatedDs := ds.DeepCopy()
		pSpec := &mutatedDs.Spec.Template.Spec
		pSpec.Containers[0].Name = tc.cntName
		UpdaterDaemonSet(mutatedDs, options.DaemonSet{
			PullIfNotPresent: tc.pullIfNotPresent,
			PFPEnable:        tc.pfpEnable,
			NodeSelector:     tc.nodeSelector,
		})
		if tc.cntName == manifests.ContainerNameNFDTopologyUpdater {
			if pSpec.Containers[0].ImagePullPolicy != pullPolicy(tc.pullIfNotPresent) {
				t.Errorf("expected container ImagePullPolicy to be: %q; got: %q", pullPolicy(tc.pullIfNotPresent), pSpec.Containers[0].ImagePullPolicy)
			}
			if !cmp.Equal(pSpec.NodeSelector, tc.nodeSelector.MatchLabels) {
				t.Errorf("expected NodeSelector to be: %v; got: %v", tc.nodeSelector.MatchLabels, pSpec.NodeSelector)
			}

			for _, arg := range tc.expectedCommandArgs {
				if !matchArgs(pSpec.Containers[0].Args, arg) {
					t.Fatalf("the container args does not container argument %q: {%v}", arg, pSpec.Containers[0].Args)
				}
			}

		} else {
			if pSpec.Containers[0].ImagePullPolicy != "" {
				t.Errorf("container name is other than %q, no changes to container are expected", manifests.ContainerNameNFDTopologyUpdater)
			}
		}
	}
}

func matchArgs(got []string, expArg string) bool {
	for _, gotArg := range got {
		if gotArg == expArg {
			return true
		}
	}
	return false
}

func pullPolicy(pullIfNotPresent bool) corev1.PullPolicy {
	if pullIfNotPresent {
		return corev1.PullIfNotPresent
	}
	return corev1.PullAlways
}
