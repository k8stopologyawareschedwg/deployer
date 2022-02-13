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
 * Copyright 2022 Red Hat, Inc.
 */

package validator

import (
	"testing"
)

func TestClusterVersionValidations(t *testing.T) {
	type testCase struct {
		version  string
		expected []ValidationResult
	}

	testCases := []testCase{
		{
			version:  "1.23",
			expected: []ValidationResult{},
		},
		{
			version: "",
			expected: []ValidationResult{
				{
					Area:      AreaCluster,
					Component: ComponentAPIVersion,
				},
			},
		},
		{
			version: "INVALID",
			expected: []ValidationResult{
				{
					Area:      AreaCluster,
					Component: ComponentAPIVersion,
				},
			},
		},
		{
			version: "1.10",
			expected: []ValidationResult{
				{
					Area:      AreaCluster,
					Component: ComponentAPIVersion,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.version, func(t *testing.T) {
			got := ValidateClusterVersion(tc.version)
			if !matchValidationResults(tc.expected, got) {
				t.Fatalf("validation failed:\nexpected=%#v\ngot=%#v", tc.expected, got)
			}
		})
	}
}
