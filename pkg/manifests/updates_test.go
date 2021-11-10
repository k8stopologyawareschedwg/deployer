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
	"reflect"
	"testing"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
)

func TestUpdateResourceTopologyExporterCommandMultipleCalls(t *testing.T) {
	type testCase struct {
		name     string
		args     []string
		vars     map[string]string
		plat     platform.Platform
		expected []string
	}

	testCases := []testCase{
		{
			name:     "kubernetes, no vars",
			args:     []string{"/bin/k8sfoo", "-v=2", "--bar=42"},
			plat:     platform.Kubernetes,
			expected: []string{"/bin/k8sfoo", "-v=2", "--bar=42", "--kubelet-config-file=/host-var/lib/kubelet/config.yaml"},
		},
		{
			name:     "openshift, no vars",
			args:     []string{"/bin/ocpfoo", "-v=3", "--baz=42"},
			plat:     platform.OpenShift,
			expected: []string{"/bin/ocpfoo", "-v=3", "--baz=42", "--topology-manager-policy=single-numa-node"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// the most important magic numbers are: 0, 1, 2, ... N
			// we avoid 0 for obvious reasons, we limit N=10 for practicalty.
			// Hence: 1, 2, N=10
			iterations := []int{1, 2, 10}
			for _, its := range iterations {
				args := append([]string{}, tc.args...)
				var retArgs []string
				for idx := 0; idx < its; idx++ {
					retArgs = UpdateResourceTopologyExporterCommand(args, tc.vars, tc.plat)
					args = retArgs
				}

				if !reflect.DeepEqual(tc.expected, retArgs) {
					t.Errorf("testcase %q iterations %d expected %v got %v", tc.name, its, tc.expected, retArgs)
				}
			}
		})
	}
}
