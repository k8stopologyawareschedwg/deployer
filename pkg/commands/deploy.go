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
 * Copyright 2021 Red Hat, Inc.
 */

package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deploy"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/api"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform/detect"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/sched"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/updaters"
	"github.com/k8stopologyawareschedwg/deployer/pkg/options"
)

func NewDeployCommand(env *deployer.Environment, commonOpts *options.Options) *cobra.Command {
	deploy := &cobra.Command{
		Use:   "deploy",
		Short: "deploy the components and configurations needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			return deploy.OnCluster(env, commonOpts)
		},
		Args: cobra.NoArgs,
	}
	deploy.PersistentFlags().BoolVarP(&commonOpts.WaitCompletion, "wait", "W", false, "wait for deployment to be all completed.")
	deploy.AddCommand(NewDeployAPICommand(env, commonOpts))
	deploy.AddCommand(NewDeploySchedulerPluginCommand(env, commonOpts))
	deploy.AddCommand(NewDeployTopologyUpdaterCommand(env, commonOpts))
	return deploy
}

func NewDeployAPICommand(env *deployer.Environment, commonOpts *options.Options) *cobra.Command {
	deploy := &cobra.Command{
		Use:   "api",
		Short: "deploy the APIs needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
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
			if err := api.Deploy(env, options.API{Platform: commonOpts.ClusterPlatform}); err != nil {
				return err
			}
			return nil
		},
		Args: cobra.NoArgs,
	}
	return deploy
}

func NewDeploySchedulerPluginCommand(env *deployer.Environment, commonOpts *options.Options) *cobra.Command {
	deploy := &cobra.Command{
		Use:   "scheduler-plugin",
		Short: "deploy the scheduler plugin needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			if err = env.EnsureClient(); err != nil {
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
			return sched.Deploy(env, options.Scheduler{
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
			})
		},
		Args: cobra.NoArgs,
	}
	return deploy
}

func NewDeployTopologyUpdaterCommand(env *deployer.Environment, commonOpts *options.Options) *cobra.Command {
	deploy := &cobra.Command{
		Use:   "topology-updater",
		Short: "deploy the topology updater needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			if err = env.EnsureClient(); err != nil {
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
			return updaters.Deploy(env, commonOpts.UpdaterType, options.Updater{
				Platform:        commonOpts.ClusterPlatform,
				PlatformVersion: commonOpts.ClusterVersion,
				WaitCompletion:  commonOpts.WaitCompletion,
				RTEConfigData:   commonOpts.RTEConfigData,
				DaemonSet:       options.ForDaemonSet(commonOpts),
				EnableCRIHooks:  commonOpts.UpdaterCRIHooksEnable,
			})
		},
		Args: cobra.NoArgs,
	}
	return deploy
}
