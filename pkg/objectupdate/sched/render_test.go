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

	"github.com/google/go-cmp/cmp"
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
			expected: `apiVersion: kubescheduler.config.k8s.io/v1beta3
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
			expected: `apiVersion: kubescheduler.config.k8s.io/v1beta3
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
			expected: `apiVersion: kubescheduler.config.k8s.io/v1beta3
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
			expected: `apiVersion: kubescheduler.config.k8s.io/v1beta3
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
		{
			name: "all params updated from empty",
			params: &manifests.ConfigParams{
				Cache: &manifests.ConfigCacheParams{
					ResyncPeriodSeconds:   newInt64(42),
					ResyncMethod:          newString("OnlyExclusiveResources"),
					ForeignPodsDetectMode: newString("OnlyExclusiveResources"),
					InformerMode:          newString("Dedicated"),
				},
			},
			initial: configTemplateEmpty,
			expected: `apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args:
      cache:
        foreignPodsDetect: OnlyExclusiveResources
        informerMode: Dedicated
        resyncMethod: OnlyExclusiveResources
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
			name: "cache params updated from empty",
			params: &manifests.ConfigParams{
				Cache: &manifests.ConfigCacheParams{
					ResyncPeriodSeconds:   newInt64(11),
					ResyncMethod:          newString("OnlyExclusiveResources"),
					ForeignPodsDetectMode: newString("OnlyExclusiveResources"),
				},
			},
			initial: configTemplateAllValues,
			expected: `apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args:
      cache:
        foreignPodsDetect: OnlyExclusiveResources
        resyncMethod: OnlyExclusiveResources
      cacheResyncPeriodSeconds: 11
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
			name: "all params updated from nonempty",
			params: &manifests.ConfigParams{
				Cache: &manifests.ConfigCacheParams{
					ResyncPeriodSeconds:   newInt64(7),
					ResyncMethod:          newString("Autodetect"),
					ForeignPodsDetectMode: newString("None"),
					InformerMode:          newString("Shared"),
				},
			},
			initial: configTemplateAllValuesFineTuned,
			expected: `apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args:
      cache:
        foreignPodsDetect: None
        informerMode: Shared
        resyncMethod: Autodetect
      cacheResyncPeriodSeconds: 7
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
			name: "partial cache params updated from empty",
			params: &manifests.ConfigParams{
				Cache: &manifests.ConfigCacheParams{
					ResyncPeriodSeconds:   newInt64(42),
					ForeignPodsDetectMode: newString("None"),
				},
			},
			initial: configTemplateEmpty,
			expected: `apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args:
      cache:
        foreignPodsDetect: None
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
			name: "partial params updated from nonempty",
			params: &manifests.ConfigParams{
				Cache: &manifests.ConfigCacheParams{
					ForeignPodsDetectMode: newString("All"),
				},
			},
			initial: configTemplateAllValuesFineTuned,
			expected: `apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args:
      cache:
        foreignPodsDetect: All
        informerMode: Dedicated
        resyncMethod: OnlyExclusiveResources
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
`,
			expectedUpdate: true,
		},
		{
			name: "partial scoring strategy params updated from empty",
			params: &manifests.ConfigParams{
				ScoringStrategy: &manifests.ScoringStrategyParams{
					Type: manifests.ScoringStrategyBalancedAllocation,
				},
			},
			initial: configTemplateEmpty,
			expected: `apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args:
      scoringStrategy:
        type: BalancedAllocation
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
			name: "partial scoring strategy params updated from empty - 2",
			params: &manifests.ConfigParams{
				ScoringStrategy: &manifests.ScoringStrategyParams{
					Type: manifests.ScoringStrategyBalancedAllocation,
					Resources: []manifests.ResourceSpecParams{
						{
							Name:   "cpu",
							Weight: int64(20),
						},
						{
							Name:   "fancy.com/device",
							Weight: int64(100),
						},
					},
				},
			},
			initial: configTemplateEmpty,
			expected: `apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args:
      scoringStrategy:
        resources:
        - name: cpu
          weight: 20
        - name: fancy.com/device
          weight: 100
        type: BalancedAllocation
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
			name: "partial scoring strategy params updated from nonempty",
			params: &manifests.ConfigParams{
				ScoringStrategy: &manifests.ScoringStrategyParams{
					Type: manifests.ScoringStrategyBalancedAllocation,
					Resources: []manifests.ResourceSpecParams{
						{
							Name:   "cpu",
							Weight: int64(20),
						},
						{
							Name:   "fancy.com/device",
							Weight: int64(100),
						},
					},
				},
			},
			initial: configTemplateAllValuesScoringFineTuned,
			expected: `apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args:
      cache:
        foreignPodsDetect: OnlyExclusiveResources
        informerMode: Dedicated
        resyncMethod: OnlyExclusiveResources
      cacheResyncPeriodSeconds: 5
      scoringStrategy:
        resources:
        - name: cpu
          weight: 20
        - name: fancy.com/device
          weight: 100
        type: BalancedAllocation
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
			name: "all params including leader election updated from empty",
			params: &manifests.ConfigParams{
				Cache: &manifests.ConfigCacheParams{
					ResyncPeriodSeconds:   newInt64(7),
					ResyncMethod:          newString("Autodetect"),
					ForeignPodsDetectMode: newString("None"),
					InformerMode:          newString("Shared"),
				},
				ScoringStrategy: &manifests.ScoringStrategyParams{
					Type: manifests.ScoringStrategyBalancedAllocation,
					Resources: []manifests.ResourceSpecParams{
						{
							Name:   "cpu",
							Weight: int64(20),
						},
						{
							Name:   "fancy.com/device",
							Weight: int64(100),
						},
					},
				},
				LeaderElection: &manifests.LeaderElectionParams{
					LeaderElect:       true,
					ResourceNamespace: "numaresources",
					ResourceName:      "nrtmatch-scheduler",
				},
			},
			initial: configTemplateEmpty,
			expected: `apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: true
  resourceName: nrtmatch-scheduler
  resourceNamespace: numaresources
profiles:
- pluginConfig:
  - args:
      cache:
        foreignPodsDetect: None
        informerMode: Shared
        resyncMethod: Autodetect
      cacheResyncPeriodSeconds: 7
      scoringStrategy:
        resources:
        - name: cpu
          weight: 20
        - name: fancy.com/device
          weight: 100
        type: BalancedAllocation
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
			name: "leader election updated from empty",
			params: &manifests.ConfigParams{
				Cache: &manifests.ConfigCacheParams{
					ResyncPeriodSeconds: newInt64(11),
				},
				LeaderElection: &manifests.LeaderElectionParams{
					LeaderElect:       true,
					ResourceNamespace: "numaresources",
					ResourceName:      "nrtmatch-scheduler",
				},
			},
			initial: configTemplateEmpty,
			expected: `apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: true
  resourceName: nrtmatch-scheduler
  resourceNamespace: numaresources
profiles:
- pluginConfig:
  - args:
      cacheResyncPeriodSeconds: 11
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
			name: "all params including leader election updated from nonempty",
			params: &manifests.ConfigParams{
				Cache: &manifests.ConfigCacheParams{
					ResyncPeriodSeconds:   newInt64(7),
					ResyncMethod:          newString("Autodetect"),
					ForeignPodsDetectMode: newString("None"),
					InformerMode:          newString("Shared"),
				},
				ScoringStrategy: &manifests.ScoringStrategyParams{
					Type: manifests.ScoringStrategyBalancedAllocation,
					Resources: []manifests.ResourceSpecParams{
						{
							Name:   "cpu",
							Weight: int64(20),
						},
						{
							Name:   "fancy.com/device",
							Weight: int64(100),
						},
					},
				},
				LeaderElection: &manifests.LeaderElectionParams{
					LeaderElect:       true,
					ResourceNamespace: "numaresources",
					ResourceName:      "nrtmatch-scheduler",
				},
			},
			initial: configTemplateAllValuesScoringFineTunedLeaderElect,
			expected: `apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: true
  resourceName: nrtmatch-scheduler
  resourceNamespace: numaresources
profiles:
- pluginConfig:
  - args:
      cache:
        foreignPodsDetect: None
        informerMode: Shared
        resyncMethod: Autodetect
      cacheResyncPeriodSeconds: 7
      scoringStrategy:
        resources:
        - name: cpu
          weight: 20
        - name: fancy.com/device
          weight: 100
        type: BalancedAllocation
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
				t.Errorf("rendering failed.\nrendered=[%s]\nexpected=[%s]\ndiff=[%s]\n", rendered, tc.expected, cmp.Diff(rendered, tc.expected))
			}
		})
	}
}

var configTemplateEmpty string = `apiVersion: kubescheduler.config.k8s.io/v1beta3
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

var configTemplateAllValues string = `apiVersion: kubescheduler.config.k8s.io/v1beta3
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

var configTemplateAllValuesFineTuned string = `apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args:
      cache:
        foreignPodsDetect: OnlyExclusiveResources
        informerMode: Dedicated
        resyncMethod: OnlyExclusiveResources
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

var configTemplateAllValuesMulti string = `apiVersion: kubescheduler.config.k8s.io/v1beta3
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

var configTemplateAllValuesScoringFineTuned string = `apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args:
      cache:
        foreignPodsDetect: OnlyExclusiveResources
        informerMode: Dedicated
        resyncMethod: OnlyExclusiveResources
      cacheResyncPeriodSeconds: 5
      scoringStrategy:
        resources:
        - name: cpu
          weight: 2	
        type: MostAllocated            
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

var configTemplateAllValuesScoringFineTunedLeaderElect string = `apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: true
  resourceNamespace: numaresources
  resourceName: nrmatch-scheduler
profiles:
- pluginConfig:
  - args:
      cache:
        foreignPodsDetect: OnlyExclusiveResources
        informerMode: Dedicated
        resyncMethod: OnlyExclusiveResources
      cacheResyncPeriodSeconds: 5
      scoringStrategy:
        resources:
        - name: cpu
          weight: 2
        type: MostAllocated
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

var configTemplateAllValuesMultiRenamed string = `apiVersion: kubescheduler.config.k8s.io/v1beta3
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

func newString(value string) *string {
	return &value
}
