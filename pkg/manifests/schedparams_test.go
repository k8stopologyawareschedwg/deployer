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
			name:           "nil",
			data:           nil,
			schedulerName:  "",
			expectedParams: ConfigParams{},
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
			schedulerName:  "topo-aware-scheduler",
			expectedParams: ConfigParams{},
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
			schedulerName:  "topology-aware-scheduler",
			expectedParams: ConfigParams{},
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
				ProfileName: "topology-aware-scheduler",
				Cache:       &ConfigCacheParams{},
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
			},
			expectedFound: true,
		},

		// keep this the last one
		{
			name: "nonzero resync period all cache params all scoringStrategyParams",
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

func newInt64(value int64) *int64 {
	return &value
}

func newString(value string) *string {
	return &value
}
