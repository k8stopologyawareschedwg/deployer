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

package platform

import "testing"

func TestRoudnTrip(t *testing.T) {
	type testCase struct {
		name       string
		expected   Platform
		expectedOK bool
	}

	testCases := []testCase{
		{
			name:       "Kubernetes",
			expected:   Kubernetes,
			expectedOK: true,
		},
		{
			name:       "OpenShift",
			expected:   OpenShift,
			expectedOK: true,
		},
		{
			name:       "HyperShift",
			expected:   HyperShift,
			expectedOK: true,
		},
		{
			name:       "foobar",
			expected:   Unknown,
			expectedOK: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ParsePlatform(tc.name)
			if ok != tc.expectedOK {
				t.Errorf("%q: got OK=%v expected=%v", tc.name, ok, tc.expectedOK)
			}
			if got != tc.expected {
				t.Errorf("%q: got=%q expected=%v", tc.name, got, tc.expected)
			}
			if tc.expectedOK {
				if got.String() != tc.name {
					t.Errorf("%q: strinq: got=%v expected=%v", tc.name, got.String(), tc.name)
				}
			}
		})
	}
}
