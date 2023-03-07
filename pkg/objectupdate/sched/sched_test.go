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
	"time"

	pluginconfig "sigs.k8s.io/scheduler-plugins/apis/config"

	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
)

func TestRenderConfig(t *testing.T) {
	rendered, err := RenderConfig(configTemplate, "test-sched-name", 42*time.Second)
	if err != nil {
		t.Errorf("RenderConfig() failed: %v", err)
	}

	schedCfg, err := manifests.DecodeSchedulerConfigFromData([]byte(rendered))
	if err != nil {
		t.Errorf("failed to decode rendered data: %v", err)
	}

	schedProf, pluginConf := findKubeSchedulerProfileByName(schedCfg, schedulerPluginName)
	if schedProf == nil || pluginConf == nil {
		t.Errorf("no profile or plugin configuration found for %q", schedulerPluginName)
	}

	confObj := pluginConf.Args.DeepCopyObject()
	pluginCfg, ok := confObj.(*pluginconfig.NodeResourceTopologyMatchArgs)
	if !ok {
		t.Errorf("unsupported plugin config type: %T", confObj)
	}

	if schedProf.SchedulerName != "test-sched-name" {
		t.Errorf("unexpected rendered data: scheduler profile name: %q", schedProf.SchedulerName)
	}

	if pluginCfg.CacheResyncPeriodSeconds != int64(42) {
		t.Errorf("unexpected rendered data: resync period: %d", pluginCfg.CacheResyncPeriodSeconds)
	}
}

var configTemplate string = `apiVersion: kubescheduler.config.k8s.io/v1beta2
kind: KubeSchedulerConfiguration
leaderElection:
  leaderElect: false
profiles:
- schedulerName: topology-aware-scheduler
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
  # optional plugin configs
  pluginConfig:
  - name: NodeResourceTopologyMatch
    args: {}`
