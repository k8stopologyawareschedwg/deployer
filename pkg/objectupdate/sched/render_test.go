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

package sched

import (
	"testing"

	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
)

func TestRenderConfig(t *testing.T) {
	type testCase struct {
		name           string
		params         *manifests.ConfigParams
		schedulerName  string
		initial        string
		expected       string
		expectedUpdate bool
	}
	testCases := []testCase{
		{
			name:     "nil",
			params:   nil,
			initial:  configTemplateEmpty,
			expected: configTemplateEmpty,
		},
		{
			name:     "nil cache",
			params:   &manifests.ConfigParams{},
			initial:  configTemplateEmpty,
			expected: configTemplateEmpty,
		},
		{
			name: "resync=zero",
			params: &manifests.ConfigParams{
				Cache: &manifests.ConfigCacheParams{
					ResyncPeriodSeconds: newInt64(0),
				},
			},
			initial:        configTemplateEmpty,
			expected:       configTemplateEmpty,
			expectedUpdate: true,
		},
		{
			name: "resync cleared if zero",
			params: &manifests.ConfigParams{
				Cache: &manifests.ConfigCacheParams{
					ResyncPeriodSeconds: newInt64(0),
				},
			},
			initial:        configTemplateAllValues,
			expected:       configTemplateEmpty,
			expectedUpdate: true,
		},
		{
			name: "resync updated from non empty",
			params: &manifests.ConfigParams{
				Cache: &manifests.ConfigCacheParams{
					ResyncPeriodSeconds: newInt64(42),
				},
			},
			initial: configTemplateAllValues,
			expected: `apiVersion: kubescheduler.config.k8s.io/v1beta2
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args:
      cacheResyncPeriodSeconds: 42
    name: NodeResourceTopologyMatch
  plugins:
    filter:
      enabled:
      - name: NodeResourceTopologyMatch
    reserve:
      enabled:
      - name: NodeResourceTopologyMatch
    score:
      enabled:
      - name: NodeResourceTopologyMatch
  schedulerName: test-sched-name
`,
			expectedUpdate: true,
		},
		{
			name: "resync updated from empty",
			params: &manifests.ConfigParams{
				Cache: &manifests.ConfigCacheParams{
					ResyncPeriodSeconds: newInt64(42),
				},
			},
			initial: configTemplateEmpty,
			expected: `apiVersion: kubescheduler.config.k8s.io/v1beta2
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args:
      cacheResyncPeriodSeconds: 42
    name: NodeResourceTopologyMatch
  plugins:
    filter:
      enabled:
      - name: NodeResourceTopologyMatch
    reserve:
      enabled:
      - name: NodeResourceTopologyMatch
    score:
      enabled:
      - name: NodeResourceTopologyMatch
  schedulerName: test-sched-name
`,
			expectedUpdate: true,
		},
		{
			name: "cannot update bad schedulerName",
			params: &manifests.ConfigParams{
				Cache: &manifests.ConfigCacheParams{
					ResyncPeriodSeconds: newInt64(42),
				},
			},
			schedulerName: "numa-aware-sched",
			initial:       configTemplateAllValues,
			expected:      configTemplateAllValues,
		},
		{
			name: "rename scheduler schedulerName multi",
			params: &manifests.ConfigParams{
				ProfileName: "renamed-sched",
			},
			initial: configTemplateAllValuesMulti,
			expected: `apiVersion: kubescheduler.config.k8s.io/v1beta2
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- plugins:
    filter:
      disabled:
      - name: '*'
      enabled:
      - name: NodeResourceFit
  schedulerName: onlyResourceFit
- pluginConfig:
  - args:
      cacheResyncPeriodSeconds: 5
    name: NodeResourceTopologyMatch
  plugins:
    filter:
      enabled:
      - name: NodeResourceTopologyMatch
    reserve:
      enabled:
      - name: NodeResourceTopologyMatch
    score:
      enabled:
      - name: NodeResourceTopologyMatch
  schedulerName: renamed-sched
`,
			expectedUpdate: true,
		},
		{
			name: "resync updated from empty multi",
			params: &manifests.ConfigParams{
				Cache: &manifests.ConfigCacheParams{
					ResyncPeriodSeconds: newInt64(42),
				},
			},
			initial: configTemplateAllValuesMulti,
			expected: `apiVersion: kubescheduler.config.k8s.io/v1beta2
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- plugins:
    filter:
      disabled:
      - name: '*'
      enabled:
      - name: NodeResourceFit
  schedulerName: onlyResourceFit
- pluginConfig:
  - args:
      cacheResyncPeriodSeconds: 42
    name: NodeResourceTopologyMatch
  plugins:
    filter:
      enabled:
      - name: NodeResourceTopologyMatch
    reserve:
      enabled:
      - name: NodeResourceTopologyMatch
    score:
      enabled:
      - name: NodeResourceTopologyMatch
  schedulerName: test-sched-name
`,
			expectedUpdate: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schedulerName := "test-sched-name"
			if tc.schedulerName != "" {
				schedulerName = tc.schedulerName
			}

			data, ok, err := RenderConfig([]byte(tc.initial), schedulerName, tc.params)
			if err != nil {
				t.Errorf("RenderConfig() failed: %v", err)
			}

			if ok != tc.expectedUpdate {
				t.Errorf("updated %v expected update %v", ok, tc.expectedUpdate)
			}

			rendered := string(data)
			if rendered != tc.expected {
				t.Errorf("rendering failed.\nrendered=[%s]\nexpected=[%s]\n", rendered, tc.expected)
			}
		})
	}
}

var configTemplateEmpty string = `apiVersion: kubescheduler.config.k8s.io/v1beta2
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args: {}
    name: NodeResourceTopologyMatch
  plugins:
    filter:
      enabled:
      - name: NodeResourceTopologyMatch
    reserve:
      enabled:
      - name: NodeResourceTopologyMatch
    score:
      enabled:
      - name: NodeResourceTopologyMatch
  schedulerName: test-sched-name
`

var configTemplateAllValues string = `apiVersion: kubescheduler.config.k8s.io/v1beta2
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args:
      cacheResyncPeriodSeconds: 5
    name: NodeResourceTopologyMatch
  plugins:
    filter:
      enabled:
      - name: NodeResourceTopologyMatch
    reserve:
      enabled:
      - name: NodeResourceTopologyMatch
    score:
      enabled:
      - name: NodeResourceTopologyMatch
  schedulerName: test-sched-name
`

var configTemplateAllValuesMulti string = `apiVersion: kubescheduler.config.k8s.io/v1beta2
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- plugins:
    filter:
      disabled:
      - name: '*'
      enabled:
      - name: NodeResourceFit
  schedulerName: onlyResourceFit
- pluginConfig:
  - args:
      cacheResyncPeriodSeconds: 5
    name: NodeResourceTopologyMatch
  plugins:
    filter:
      enabled:
      - name: NodeResourceTopologyMatch
    reserve:
      enabled:
      - name: NodeResourceTopologyMatch
    score:
      enabled:
      - name: NodeResourceTopologyMatch
  schedulerName: test-sched-name
`

var configTemplateAllValuesMultiRenamed string = `apiVersion: kubescheduler.config.k8s.io/v1beta2
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- plugins:
    filter:
      disabled:
      - name: '*'
      enabled:
      - name: NodeResourceFit
  schedulerName: onlyResourceFit
- pluginConfig:
  - args:
      cacheResyncPeriodSeconds: 5
    name: NodeResourceTopologyMatch
  plugins:
    filter:
      enabled:
      - name: NodeResourceTopologyMatch
    reserve:
      enabled:
      - name: NodeResourceTopologyMatch
    score:
      enabled:
      - name: NodeResourceTopologyMatch
  schedulerName: renamed-sched
`

func newInt64(value int64) *int64 {
	return &value
}
