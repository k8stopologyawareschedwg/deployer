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
	"fmt"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/api"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform/detect"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/sched"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/updaters"
	"github.com/k8stopologyawareschedwg/deployer/pkg/options"
)

func OnCluster(env *deployer.Environment, commonOpts *options.Options) error {
	if err := env.EnsureClient(); err != nil {
		return err
	}

	platDetect, reason, _ := detect.FindPlatform(env.Ctx, commonOpts.UserPlatform)
	commonOpts.ClusterPlatform = platDetect.Discovered
	if commonOpts.ClusterPlatform == platform.Unknown {
		return fmt.Errorf("cannot autodetect the platform, and no platform given")
	}
	versionDetect, source, _ := detect.FindVersion(env.Ctx, platDetect.Discovered, commonOpts.UserPlatformVersion)
	commonOpts.ClusterVersion = versionDetect.Discovered
	if commonOpts.ClusterVersion == platform.MissingVersion {
		return fmt.Errorf("cannot autodetect the platform version, and no version given")
	}

	env.Log.Info("detection", "platform", commonOpts.ClusterPlatform, "reason", reason, "version", commonOpts.ClusterVersion, "source", source)
	if err := api.Deploy(env, options.API{
		Platform: commonOpts.ClusterPlatform,
	}); err != nil {
		return err
	}
	if err := updaters.Deploy(env, commonOpts.UpdaterType, options.Updater{
		Platform:        commonOpts.ClusterPlatform,
		PlatformVersion: commonOpts.ClusterVersion,
		WaitCompletion:  commonOpts.WaitCompletion,
		RTEConfigData:   commonOpts.RTEConfigData,
		DaemonSet:       options.ForDaemonSet(commonOpts),
		EnableCRIHooks:  commonOpts.UpdaterCRIHooksEnable,
	}); err != nil {
		return err
	}
	if err := sched.Deploy(env, options.Scheduler{
		Platform:               commonOpts.ClusterPlatform,
		WaitCompletion:         commonOpts.WaitCompletion,
		Replicas:               int32(commonOpts.Replicas),
		PullIfNotPresent:       commonOpts.PullIfNotPresent,
		ProfileName:            commonOpts.SchedProfileName,
		CacheResyncPeriod:      commonOpts.SchedResyncPeriod,
		CtrlPlaneAffinity:      commonOpts.SchedCtrlPlaneAffinity,
		Verbose:                commonOpts.SchedVerbose,
		ScoringStratConfigData: commonOpts.SchedScoringStratConfigData,
		CacheParamsConfigData:  commonOpts.SchedCacheParamsConfigData,
		LeaderElection:         commonOpts.Replicas > 1,
		LeaderElectionResource: commonOpts.SchedLeaderElectResource,
	}); err != nil {
		return err
	}
	return nil
}
