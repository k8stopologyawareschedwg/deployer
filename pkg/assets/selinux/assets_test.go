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
 * Copyright 2024 Red Hat, Inc.
 */

package selinux

import (
	"os"
	"testing"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
)

func TestGetPolicy(t *testing.T) {
	type testCase struct {
		name          string
		ver           platform.Version
		expectedError bool
	}

	testCases := []testCase{
		{
			name:          "latest", // at time of writing. Keep me updated!
			ver:           platform.Version("v4.18"),
			expectedError: false,
		},
		{
			name:          "supported",
			ver:           platform.Version("v4.12"),
			expectedError: false,
		},
		{
			name:          "more recent",
			ver:           platform.Version("v4.99"),
			expectedError: false,
		},
		{
			name:          "too old",
			ver:           platform.Version("v3.11"),
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := GetPolicy(tc.ver)
			gotErr := (err != nil)
			if gotErr != tc.expectedError {
				t.Fatalf("GetPolicy(%v) unexpected result: expected error %v got %v", tc.ver, tc.expectedError, gotErr)
			}
		})
	}
}

func TestPolicyDir(t *testing.T) {
	numOfVersions := len(knownVersions())
	dir, err := os.ReadDir(policyDir)
	if err != nil {
		t.Fatal(err)
	}
	numOfCils := len(dir)
	if numOfVersions != numOfCils {
		t.Fatalf("number of known version is different than number of cil files. knownVersions=%d,  cil files=%d", numOfVersions, numOfCils)
	}
}
