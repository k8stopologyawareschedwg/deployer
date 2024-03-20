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

package manifests

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func TestDecodeSchedulerConfigFromData(t *testing.T) {
	type testCase struct {
		name           string
		data           []byte
		schedulerName  string
		expectedFound  bool
		expectedParams ConfigParams
	}
	testCases := []testCase{
		{
			name:          "nil",
			data:          nil,
			schedulerName: "",
			expectedParams: ConfigParams{
				LeaderElection: &LeaderElectionParams{},
			},
		},
		{
			name: "bad scheduler name",
			data: []byte(`apiVersion: kubescheduler.config.k8s.io/v1beta3
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
  schedulerName: topology-aware-scheduler
`),
			schedulerName: "topo-aware-scheduler",
			expectedParams: ConfigParams{
				LeaderElection: &LeaderElectionParams{},
			},
		},
		{
			name: "bad scheduler params name",
			data: []byte(`apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args: {}
    name: noderestopo
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
  schedulerName: topology-aware-scheduler
`),
			schedulerName: "topology-aware-scheduler",
			expectedParams: ConfigParams{
				LeaderElection: &LeaderElectionParams{},
			},
		},
		{
			name: "empty params",
			data: []byte(`apiVersion: kubescheduler.config.k8s.io/v1beta3
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
  schedulerName: topology-aware-scheduler
`),
			schedulerName: "topology-aware-scheduler",
			expectedParams: ConfigParams{
				ProfileName:    "topology-aware-scheduler",
				Cache:          &ConfigCacheParams{},
				LeaderElection: &LeaderElectionParams{},
			},
			expectedFound: true,
		},
		{
			name: "nonzero resync period",
			data: []byte(`apiVersion: kubescheduler.config.k8s.io/v1beta3
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
  schedulerName: topology-aware-scheduler
`),
			schedulerName: "topology-aware-scheduler",
			expectedParams: ConfigParams{
				ProfileName: "topology-aware-scheduler",
				Cache: &ConfigCacheParams{
					ResyncPeriodSeconds: newInt64(5),
				},
				LeaderElection: &LeaderElectionParams{},
			},
			expectedFound: true,
		},
		{
			name: "nonzero resync period and all cache params",
			data: []byte(`apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args:
      cache:
        foreignPodsDetect: None
        resyncMethod: Autodetect
        informerMode: Dedicated
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
  schedulerName: topology-aware-scheduler
`),
			schedulerName: "topology-aware-scheduler",
			expectedParams: ConfigParams{
				ProfileName: "topology-aware-scheduler",
				Cache: &ConfigCacheParams{
					ResyncPeriodSeconds:   newInt64(5),
					ResyncMethod:          newString("Autodetect"),
					ForeignPodsDetectMode: newString("None"),
					InformerMode:          newString("Dedicated"),
				},
				LeaderElection: &LeaderElectionParams{},
			},
			expectedFound: true,
		},
		{
			name: "nonzero resync period and some cache params",
			data: []byte(`apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args:
      cache:
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
  schedulerName: topology-aware-scheduler
`),
			schedulerName: "topology-aware-scheduler",
			expectedParams: ConfigParams{
				ProfileName: "topology-aware-scheduler",
				Cache: &ConfigCacheParams{
					ResyncPeriodSeconds: newInt64(5),
					ResyncMethod:        newString("OnlyExclusiveResources"),
				},
				LeaderElection: &LeaderElectionParams{},
			},
			expectedFound: true,
		},
		{
			name: "zero resync period and some cache params - 2",
			data: []byte(`apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args:
      cache:
        foreignPodsDetect: OnlyExclusiveResources
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
  schedulerName: topology-aware-scheduler
`),
			schedulerName: "topology-aware-scheduler",
			expectedParams: ConfigParams{
				ProfileName: "topology-aware-scheduler",
				Cache: &ConfigCacheParams{
					ForeignPodsDetectMode: newString("OnlyExclusiveResources"),
				},
				LeaderElection: &LeaderElectionParams{},
			},
			expectedFound: true,
		},
		{
			name: "zero resync period and some cache params - 3",
			data: []byte(`apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args:
      cache:
        informerMode: Shared
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
  schedulerName: topology-aware-scheduler
`),
			schedulerName: "topology-aware-scheduler",
			expectedParams: ConfigParams{
				ProfileName: "topology-aware-scheduler",
				Cache: &ConfigCacheParams{
					InformerMode: newString("Shared"),
				},
				LeaderElection: &LeaderElectionParams{},
			},
			expectedFound: true,
		},
		{
			name: "all scoringStrategy params",
			data: []byte(`apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args:
      scoringStrategy:
        type: MostAllocated
        resources:
        - name: cpu
          weight: 10
        - name: memory
          weight: 5
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
  schedulerName: topology-aware-scheduler
`),
			schedulerName: "topology-aware-scheduler",
			expectedParams: ConfigParams{
				ProfileName: "topology-aware-scheduler",
				Cache:       &ConfigCacheParams{},
				ScoringStrategy: &ScoringStrategyParams{
					Type: "MostAllocated",
					Resources: []ResourceSpecParams{
						{
							Name:   "cpu",
							Weight: int64(10),
						},
						{
							Name:   "memory",
							Weight: int64(5),
						},
					},
				},
				LeaderElection: &LeaderElectionParams{},
			},
			expectedFound: true,
		},
		{
			name: "some scoringStrategy params - 1",
			data: []byte(`apiVersion: kubescheduler.config.k8s.io/v1beta3
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
  schedulerName: topology-aware-scheduler
`),
			schedulerName: "topology-aware-scheduler",
			expectedParams: ConfigParams{
				ProfileName: "topology-aware-scheduler",
				Cache:       &ConfigCacheParams{},
				ScoringStrategy: &ScoringStrategyParams{
					Type: "BalancedAllocation",
				},
				LeaderElection: &LeaderElectionParams{},
			},
			expectedFound: true,
		},
		{
			name: "some scoringStrategy params - 2",
			data: []byte(`apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- pluginConfig:
  - args:
      scoringStrategy:
        resources:
        - name: device.io/foobar
          weight: 100
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
  schedulerName: topology-aware-scheduler
`),
			schedulerName: "topology-aware-scheduler",
			expectedParams: ConfigParams{
				ProfileName: "topology-aware-scheduler",
				Cache:       &ConfigCacheParams{},
				ScoringStrategy: &ScoringStrategyParams{
					Resources: []ResourceSpecParams{
						{
							Name:   "device.io/foobar",
							Weight: int64(100),
						},
					},
				},
				LeaderElection: &LeaderElectionParams{},
			},
			expectedFound: true,
		},
		{
			name: "minimal leader election params",
			data: []byte(`apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: true
profiles:
- pluginConfig:
  - args:
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
  schedulerName: topology-aware-scheduler-leader-elect
`),
			schedulerName: "topology-aware-scheduler-leader-elect",
			expectedParams: ConfigParams{
				ProfileName: "topology-aware-scheduler-leader-elect",
				Cache: &ConfigCacheParams{
					ResyncPeriodSeconds: newInt64(7),
				},
				LeaderElection: &LeaderElectionParams{
					LeaderElect: true,
				},
			},
			expectedFound: true,
		},
		{
			name: "partial leader election params - 1",
			data: []byte(`apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: true
  resourceNamespace: numa-aware-sched
profiles:
- pluginConfig:
  - args:
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
  schedulerName: topology-aware-scheduler-leader-elect
`),
			schedulerName: "topology-aware-scheduler-leader-elect",
			expectedParams: ConfigParams{
				ProfileName: "topology-aware-scheduler-leader-elect",
				Cache: &ConfigCacheParams{
					ResyncPeriodSeconds: newInt64(7),
				},
				LeaderElection: &LeaderElectionParams{
					LeaderElect:       true,
					ResourceNamespace: "numa-aware-sched",
				},
			},
			expectedFound: true,
		},
		{
			name: "partial leader election params - 2",
			data: []byte(`apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: true
  resourceName: numa-nrtmatch-sched
profiles:
- pluginConfig:
  - args:
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
  schedulerName: topology-aware-scheduler-leader-elect
`),
			schedulerName: "topology-aware-scheduler-leader-elect",
			expectedParams: ConfigParams{
				ProfileName: "topology-aware-scheduler-leader-elect",
				Cache: &ConfigCacheParams{
					ResyncPeriodSeconds: newInt64(7),
				},
				LeaderElection: &LeaderElectionParams{
					LeaderElect:  true,
					ResourceName: "numa-nrtmatch-sched",
				},
			},
			expectedFound: true,
		},
		{
			name: "full leader election params",
			data: []byte(`apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: true
  resourceNamespace: numa-aware-sched
  resourceName: numa-nrtmatch-sched
profiles:
- pluginConfig:
  - args:
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
  schedulerName: topology-aware-scheduler-leader-elect
`),
			schedulerName: "topology-aware-scheduler-leader-elect",
			expectedParams: ConfigParams{
				ProfileName: "topology-aware-scheduler-leader-elect",
				Cache: &ConfigCacheParams{
					ResyncPeriodSeconds: newInt64(7),
				},
				LeaderElection: &LeaderElectionParams{
					LeaderElect:       true,
					ResourceNamespace: "numa-aware-sched",
					ResourceName:      "numa-nrtmatch-sched",
				},
			},
			expectedFound: true,
		},
		// keep this the last one
		{
			name: "nonzero resync period all params",
			data: []byte(`apiVersion: kubescheduler.config.k8s.io/v1beta3
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: true
  resourceNamespace: numa-aware-sched
  resourceName: numa-nrtmatch-sched
profiles:
- pluginConfig:
  - args:
      cache:
        foreignPodsDetect: None
        resyncMethod: Autodetect
        informerMode: Dedicated
      cacheResyncPeriodSeconds: 5
      scoringStrategy:
        type: BalancedAllocation
        resources:
        - name: cpu
          weight: 10
        - name: memory
          weight: 5
        - name: device.io/foobar
          weight: 20
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
  schedulerName: topology-aware-scheduler
`),
			schedulerName: "topology-aware-scheduler",
			expectedParams: ConfigParams{
				ProfileName: "topology-aware-scheduler",
				Cache: &ConfigCacheParams{
					ResyncPeriodSeconds:   newInt64(5),
					ResyncMethod:          newString("Autodetect"),
					ForeignPodsDetectMode: newString("None"),
					InformerMode:          newString("Dedicated"),
				},
				ScoringStrategy: &ScoringStrategyParams{
					Type: "BalancedAllocation",
					Resources: []ResourceSpecParams{
						{
							Name:   "cpu",
							Weight: int64(10),
						},
						{
							Name:   "memory",
							Weight: int64(5),
						},
						{
							Name:   "device.io/foobar",
							Weight: int64(20),
						},
					},
				},
				LeaderElection: &LeaderElectionParams{
					LeaderElect:       true,
					ResourceNamespace: "numa-aware-sched",
					ResourceName:      "numa-nrtmatch-sched",
				},
			},
			expectedFound: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			allParams, err := DecodeSchedulerProfilesFromData(tc.data)
			if err != nil {
				t.Fatalf("unexpected error [%v]", err)
			}
			if !tc.expectedFound {
				return // nothing else to do
			}

			if len(allParams) != 1 {
				t.Fatalf("unexpected params: found %d", len(allParams))
			}
			params := FindSchedulerProfileByName(allParams, tc.schedulerName)
			if params == nil {
				t.Fatalf("cannot find params for %q", tc.schedulerName)
			}

			if !reflect.DeepEqual(params, &tc.expectedParams) {
				t.Fatalf("params got %s expected %s", toJSON(params), toJSON(tc.expectedParams))
			}
		})
	}
}

func toJSON(v any) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("<err=%v>", err)
	}
	return string(data)
}
