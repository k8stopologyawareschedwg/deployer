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

package deploy

import (
	"time"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
)

type Options struct {
	UserPlatform           platform.Platform
	UserPlatformVersion    platform.Version
	Replicas               int
	RTEConfigData          string
	PullIfNotPresent       bool
	UpdaterType            string
	UpdaterPFPEnable       bool
	UpdaterNotifEnable     bool
	UpdaterCRIHooksEnable  bool
	UpdaterSyncPeriod      time.Duration
	UpdaterVerbose         int
	SchedProfileName       string
	SchedResyncPeriod      time.Duration
	SchedVerbose           int
	SchedCtrlPlaneAffinity bool
	WaitInterval           time.Duration
	WaitTimeout            time.Duration
	ClusterPlatform        platform.Platform
	ClusterVersion         platform.Version
	WaitCompletion         bool
}
