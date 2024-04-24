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

package images

import (
	"bytes"
	"strings"
	"testing"
)

func TestOutputBasics(t *testing.T) {
	imgs := GetWithFunc(false, testGetImage)

	imo := NewOutput(imgs, "foobar")
	images := imo.ToList()
	if len(images) != 3 {
		t.Errorf("unexpected image list content: %#v", images)
	}
}

func TestOutput(t *testing.T) {
	type testCase struct {
		name        string
		kind        int
		updaterType string
		expected    string
	}

	testCases := []testCase{
		{
			name:        "text/rte",
			kind:        FormatText,
			updaterType: "RTE",
			expected:    "TAS_SCHEDULER_PLUGIN_IMAGE=sched_sched\nTAS_SCHEDULER_PLUGIN_CONTROLLER_IMAGE=sched_ctrl\nTAS_RESOURCE_EXPORTER_IMAGE=rte",
		},
		{
			name:        "text/nfd",
			kind:        FormatText,
			updaterType: "NFD",
			expected:    "TAS_SCHEDULER_PLUGIN_IMAGE=sched_sched\nTAS_SCHEDULER_PLUGIN_CONTROLLER_IMAGE=sched_ctrl\nTAS_RESOURCE_EXPORTER_IMAGE=nfd",
		},
		{
			name:        "json/rte",
			kind:        FormatJSON,
			updaterType: "RTE",
			expected:    `{"topology_updater":"rte","scheduler_plugin":"sched_sched","scheduler_controller":"sched_ctrl"}`,
		},
		{
			name:        "json/nfd",
			kind:        FormatJSON,
			updaterType: "NFD",
			expected:    `{"topology_updater":"nfd","scheduler_plugin":"sched_sched","scheduler_controller":"sched_ctrl"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			imgs := GetWithFunc(false, testGetImage)

			imo := NewOutput(imgs, tc.updaterType)
			var buf bytes.Buffer
			imo.Format(tc.kind, &buf)
			got := strings.TrimSpace(buf.String())
			if got != tc.expected {
				t.Errorf("%s: got=%q expected=%q", tc.name, got, tc.expected)
			}
		})
	}
}

func testGetImage(key string) (string, bool) {
	switch key {
	case "TAS_SCHEDULER_PLUGIN_IMAGE":
		return "sched_sched", true
	case "TAS_SCHEDULER_PLUGIN_CONTROLLER_IMAGE":
		return "sched_ctrl", true
	case "TAS_RESOURCE_EXPORTER_IMAGE":
		return "rte", true
	case "TAS_NODE_FEATURE_DISCOVERY_IMAGE":
		return "nfd", true
	default:
		return "", false
	}
}
