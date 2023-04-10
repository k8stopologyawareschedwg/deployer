/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubeletconfig

import (
	"bytes"
	"testing"
)

func TestFindProxyPort(t *testing.T) {
	type testCase struct {
		name         string
		text         string
		expectedPort int
		expectedErr  bool
	}

	testCases := []testCase{
		{
			name:         "empty",
			text:         "",
			expectedPort: -1,
			expectedErr:  true,
		},
		{
			name:         "no match",
			text:         "foo\nbar\nbaz",
			expectedPort: -1,
			expectedErr:  true,
		},
		{
			name:         "trivial match",
			text:         "Starting to serve on 127.0.0.1:12345\n",
			expectedPort: 12345,
			expectedErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.text)
			port, err := FindProxyPort(buf)
			gotErr := err != nil
			if gotErr != tc.expectedErr {
				t.Errorf("error: got=%v expected=%v", gotErr, tc.expectedErr)
			}
			if port != tc.expectedPort {
				t.Errorf("port: got=%v expected=%v", port, tc.expectedPort)
			}
		})
	}
}
