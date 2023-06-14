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

	"sigs.k8s.io/yaml"

	corev1 "k8s.io/api/core/v1"
)

func TestSetPodSchedulerAffinityOnControlPlane(t *testing.T) {

	type testCase struct {
		name         string
		podSpec      *corev1.PodSpec
		expectedYAML string
	}

	testCases := []testCase{
		{
			name:         "nil",
			expectedYAML: "null\n",
		},
		{
			name:         "empty",
			podSpec:      &corev1.PodSpec{},
			expectedYAML: podSpecOnlyAffinity,
		},
		{
			name: "partial affinity",
			podSpec: &corev1.PodSpec{
				Affinity: &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{},
				},
			},
			expectedYAML: podSpecOnlyAffinity,
		},
		{
			name: "partial tolerations",
			podSpec: &corev1.PodSpec{
				Tolerations: []corev1.Toleration{
					{
						Key:    NodeRoleControlPlane,
						Effect: corev1.TaintEffectNoSchedule,
					},
				},
			},
			expectedYAML: podSpecOnlyAffinity,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.podSpec.DeepCopy()
			SetPodSchedulerAffinityOnControlPlane(got)
			data, err := yaml.Marshal(got)
			if err != nil {
				t.Errorf("error marshalling yaml: %v", err)
			}
			gotYAML := string(data)
			if gotYAML != tc.expectedYAML {
				t.Errorf("output mismatch:\ngot=%v\nexpected=%v\n", gotYAML, tc.expectedYAML)
			}
		})
	}
}

const podSpecOnlyAffinity = `affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
      - matchExpressions:
        - key: node-role.kubernetes.io/control-plane
          operator: Exists
containers: null
tolerations:
- effect: NoSchedule
  key: node-role.kubernetes.io/control-plane
- effect: NoSchedule
  key: node-role.kubernetes.io/master
`
