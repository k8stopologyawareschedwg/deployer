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

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/api"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform/detect"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/sched"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/updaters"
)

type DeployOptions struct {
	clusterPlatform platform.Platform
	clusterVersion  platform.Version
	waitCompletion  bool
}

func NewDeployCommand(env *deployer.Environment, commonOpts *CommonOptions) *cobra.Command {
	opts := &DeployOptions{}
	deploy := &cobra.Command{
		Use:   "deploy",
		Short: "deploy the components and configurations needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			return deployOnCluster(env, commonOpts, opts)
		},
		Args: cobra.NoArgs,
	}
	deploy.PersistentFlags().BoolVarP(&opts.waitCompletion, "wait", "W", false, "wait for deployment to be all completed.")
	deploy.AddCommand(NewDeployAPICommand(env, commonOpts, opts))
	deploy.AddCommand(NewDeploySchedulerPluginCommand(env, commonOpts, opts))
	deploy.AddCommand(NewDeployTopologyUpdaterCommand(env, commonOpts, opts))
	return deploy
}

func NewRemoveCommand(env *deployer.Environment, commonOpts *CommonOptions) *cobra.Command {
	opts := &DeployOptions{}
	remove := &cobra.Command{
		Use:   "remove",
		Short: "remove the components and configurations needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			if err = env.EnsureClient(); err != nil {
				return err
			}

			platDetect, reason, _ := detect.FindPlatform(env.Ctx, commonOpts.UserPlatform)
			opts.clusterPlatform = platDetect.Discovered
			if opts.clusterPlatform == platform.Unknown {
				return fmt.Errorf("cannot autodetect the platform, and no platform given")
			}
			versionDetect, source, _ := detect.FindVersion(env.Ctx, platDetect.Discovered, commonOpts.UserPlatformVersion)
			opts.clusterVersion = versionDetect.Discovered
			if opts.clusterVersion == platform.MissingVersion {
				return fmt.Errorf("cannot autodetect the platform version, and no version given")
			}
			env.Log.Info("detection", "platform", opts.clusterPlatform, "reason", reason, "version", opts.clusterVersion, "source", source)

			err = sched.Remove(env, sched.Options{
				Platform:          opts.clusterPlatform,
				WaitCompletion:    opts.waitCompletion,
				Replicas:          int32(commonOpts.Replicas),
				RTEConfigData:     commonOpts.RTEConfigData,
				PullIfNotPresent:  commonOpts.PullIfNotPresent,
				CacheResyncPeriod: commonOpts.SchedResyncPeriod,
				CtrlPlaneAffinity: commonOpts.SchedCtrlPlaneAffinity,
			})
			if err != nil {
				// intentionally keep going to remove as much as possible
				env.Log.Info("while removing", "error", err)
			}
			err = updaters.Remove(env, commonOpts.UpdaterType, updaters.Options{
				Platform:        opts.clusterPlatform,
				PlatformVersion: opts.clusterVersion,
				WaitCompletion:  opts.waitCompletion,
				RTEConfigData:   commonOpts.RTEConfigData,
				DaemonSet:       daemonSetOptionsFromCommonOptions(commonOpts),
				EnableCRIHooks:  commonOpts.UpdaterCRIHooksEnable,
			})
			if err != nil {
				// intentionally keep going to remove as much as possible
				env.Log.Info("while removing", "error", err)
			}
			err = api.Remove(env, api.Options{
				Platform: opts.clusterPlatform,
			})
			if err != nil {
				// intentionally keep going to remove as much as possible
				env.Log.Info("while removing", "error", err)
			}
			return nil
		},
		Args: cobra.NoArgs,
	}
	remove.PersistentFlags().BoolVarP(&opts.waitCompletion, "wait", "W", false, "wait for removal to be all completed.")
	remove.AddCommand(NewRemoveAPICommand(env, commonOpts, opts))
	remove.AddCommand(NewRemoveSchedulerPluginCommand(env, commonOpts, opts))
	remove.AddCommand(NewRemoveTopologyUpdaterCommand(env, commonOpts, opts))
	return remove
}

func NewDeployAPICommand(env *deployer.Environment, commonOpts *CommonOptions, opts *DeployOptions) *cobra.Command {
	deploy := &cobra.Command{
		Use:   "api",
		Short: "deploy the APIs needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := env.EnsureClient(); err != nil {
				return err
			}

			platDetect, reason, _ := detect.FindPlatform(env.Ctx, commonOpts.UserPlatform)
			opts.clusterPlatform = platDetect.Discovered
			if opts.clusterPlatform == platform.Unknown {
				return fmt.Errorf("cannot autodetect the platform, and no platform given")
			}
			versionDetect, source, _ := detect.FindVersion(env.Ctx, platDetect.Discovered, commonOpts.UserPlatformVersion)
			opts.clusterVersion = versionDetect.Discovered
			if opts.clusterVersion == platform.MissingVersion {
				return fmt.Errorf("cannot autodetect the platform version, and no version given")
			}

			env.Log.Info("detection", "platform", opts.clusterPlatform, "reason", reason, "version", opts.clusterVersion, "source", source)
			if err := api.Deploy(env, api.Options{Platform: opts.clusterPlatform}); err != nil {
				return err
			}
			return nil
		},
		Args: cobra.NoArgs,
	}
	return deploy
}

func NewDeploySchedulerPluginCommand(env *deployer.Environment, commonOpts *CommonOptions, opts *DeployOptions) *cobra.Command {
	deploy := &cobra.Command{
		Use:   "scheduler-plugin",
		Short: "deploy the scheduler plugin needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			if err = env.EnsureClient(); err != nil {
				return err
			}

			platDetect, reason, _ := detect.FindPlatform(env.Ctx, commonOpts.UserPlatform)
			opts.clusterPlatform = platDetect.Discovered
			if opts.clusterPlatform == platform.Unknown {
				return fmt.Errorf("cannot autodetect the platform, and no platform given")
			}
			versionDetect, source, _ := detect.FindVersion(env.Ctx, platDetect.Discovered, commonOpts.UserPlatformVersion)
			opts.clusterVersion = versionDetect.Discovered
			if opts.clusterVersion == platform.MissingVersion {
				return fmt.Errorf("cannot autodetect the platform version, and no version given")
			}

			env.Log.Info("detection", "platform", opts.clusterPlatform, "reason", reason, "version", opts.clusterVersion, "source", source)
			return sched.Deploy(env, sched.Options{
				Platform:          opts.clusterPlatform,
				WaitCompletion:    opts.waitCompletion,
				Replicas:          int32(commonOpts.Replicas),
				RTEConfigData:     commonOpts.RTEConfigData,
				PullIfNotPresent:  commonOpts.PullIfNotPresent,
				CacheResyncPeriod: commonOpts.SchedResyncPeriod,
				CtrlPlaneAffinity: commonOpts.SchedCtrlPlaneAffinity,
				Verbose:           commonOpts.SchedVerbose,
			})
		},
		Args: cobra.NoArgs,
	}
	return deploy
}

func NewDeployTopologyUpdaterCommand(env *deployer.Environment, commonOpts *CommonOptions, opts *DeployOptions) *cobra.Command {
	deploy := &cobra.Command{
		Use:   "topology-updater",
		Short: "deploy the topology updater needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			if err = env.EnsureClient(); err != nil {
				return err
			}

			platDetect, reason, _ := detect.FindPlatform(env.Ctx, commonOpts.UserPlatform)
			opts.clusterPlatform = platDetect.Discovered
			if opts.clusterPlatform == platform.Unknown {
				return fmt.Errorf("cannot autodetect the platform, and no platform given")
			}
			versionDetect, source, _ := detect.FindVersion(env.Ctx, platDetect.Discovered, commonOpts.UserPlatformVersion)
			opts.clusterVersion = versionDetect.Discovered
			if opts.clusterVersion == platform.MissingVersion {
				return fmt.Errorf("cannot autodetect the platform version, and no version given")
			}

			env.Log.Info("detection", "platform", opts.clusterPlatform, "reason", reason, "version", opts.clusterVersion, "source", source)
			return updaters.Deploy(env, commonOpts.UpdaterType, updaters.Options{
				Platform:        opts.clusterPlatform,
				PlatformVersion: opts.clusterVersion,
				WaitCompletion:  opts.waitCompletion,
				RTEConfigData:   commonOpts.RTEConfigData,
				DaemonSet:       daemonSetOptionsFromCommonOptions(commonOpts),
				EnableCRIHooks:  commonOpts.UpdaterCRIHooksEnable,
			})
		},
		Args: cobra.NoArgs,
	}
	return deploy
}

func NewRemoveAPICommand(env *deployer.Environment, commonOpts *CommonOptions, opts *DeployOptions) *cobra.Command {
	remove := &cobra.Command{
		Use:   "api",
		Short: "remove the APIs needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			if err = env.EnsureClient(); err != nil {
				return err
			}

			platDetect, reason, _ := detect.FindPlatform(env.Ctx, commonOpts.UserPlatform)
			opts.clusterPlatform = platDetect.Discovered
			if opts.clusterPlatform == platform.Unknown {
				return fmt.Errorf("cannot autodetect the platform, and no platform given")
			}
			versionDetect, source, _ := detect.FindVersion(env.Ctx, platDetect.Discovered, commonOpts.UserPlatformVersion)
			opts.clusterVersion = versionDetect.Discovered
			if opts.clusterVersion == platform.MissingVersion {
				return fmt.Errorf("cannot autodetect the platform version, and no version given")
			}

			env.Log.Info("detection", "platform", opts.clusterPlatform, "reason", reason, "version", opts.clusterVersion, "source", source)
			if err := api.Remove(env, api.Options{Platform: opts.clusterPlatform}); err != nil {
				return err
			}
			return nil
		},
		Args: cobra.NoArgs,
	}
	return remove
}

func NewRemoveSchedulerPluginCommand(env *deployer.Environment, commonOpts *CommonOptions, opts *DeployOptions) *cobra.Command {
	remove := &cobra.Command{
		Use:   "scheduler-plugin",
		Short: "remove the scheduler plugin needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			if err = env.EnsureClient(); err != nil {
				return err
			}

			platDetect, reason, _ := detect.FindPlatform(env.Ctx, commonOpts.UserPlatform)
			opts.clusterPlatform = platDetect.Discovered
			if opts.clusterPlatform == platform.Unknown {
				return fmt.Errorf("cannot autodetect the platform, and no platform given")
			}
			versionDetect, source, _ := detect.FindVersion(env.Ctx, platDetect.Discovered, commonOpts.UserPlatformVersion)
			opts.clusterVersion = versionDetect.Discovered
			if opts.clusterVersion == platform.MissingVersion {
				return fmt.Errorf("cannot autodetect the platform version, and no version given")
			}

			env.Log.Info("detection", "platform", opts.clusterPlatform, "reason", reason, "version", opts.clusterVersion, "source", source)
			return sched.Remove(env, sched.Options{
				Platform:          opts.clusterPlatform,
				WaitCompletion:    opts.waitCompletion,
				Replicas:          int32(commonOpts.Replicas),
				RTEConfigData:     commonOpts.RTEConfigData,
				PullIfNotPresent:  commonOpts.PullIfNotPresent,
				CacheResyncPeriod: commonOpts.SchedResyncPeriod,
				CtrlPlaneAffinity: commonOpts.SchedCtrlPlaneAffinity,
				Verbose:           commonOpts.SchedVerbose,
			})
		},
		Args: cobra.NoArgs,
	}
	return remove
}

func NewRemoveTopologyUpdaterCommand(env *deployer.Environment, commonOpts *CommonOptions, opts *DeployOptions) *cobra.Command {
	remove := &cobra.Command{
		Use:   "topology-updater",
		Short: "remove the topology updater needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			if err = env.EnsureClient(); err != nil {
				return err
			}

			platDetect, reason, _ := detect.FindPlatform(env.Ctx, commonOpts.UserPlatform)
			opts.clusterPlatform = platDetect.Discovered
			if opts.clusterPlatform == platform.Unknown {
				return fmt.Errorf("cannot autodetect the platform, and no platform given")
			}
			versionDetect, source, _ := detect.FindVersion(env.Ctx, platDetect.Discovered, commonOpts.UserPlatformVersion)
			opts.clusterVersion = versionDetect.Discovered
			if opts.clusterVersion == platform.MissingVersion {
				return fmt.Errorf("cannot autodetect the platform version, and no version given")
			}

			env.Log.Info("detection", "platform", opts.clusterPlatform, "reason", reason, "version", opts.clusterVersion, "source", source)
			return updaters.Remove(env, commonOpts.UpdaterType, updaters.Options{
				Platform:        opts.clusterPlatform,
				PlatformVersion: opts.clusterVersion,
				WaitCompletion:  opts.waitCompletion,
				RTEConfigData:   commonOpts.RTEConfigData,
				DaemonSet:       daemonSetOptionsFromCommonOptions(commonOpts),
				EnableCRIHooks:  commonOpts.UpdaterCRIHooksEnable,
			})
		},
		Args: cobra.NoArgs,
	}
	return remove
}

func deployOnCluster(env *deployer.Environment, commonOpts *CommonOptions, opts *DeployOptions) error {
	if err := env.EnsureClient(); err != nil {
		return err
	}

	platDetect, reason, _ := detect.FindPlatform(env.Ctx, commonOpts.UserPlatform)
	opts.clusterPlatform = platDetect.Discovered
	if opts.clusterPlatform == platform.Unknown {
		return fmt.Errorf("cannot autodetect the platform, and no platform given")
	}
	versionDetect, source, _ := detect.FindVersion(env.Ctx, platDetect.Discovered, commonOpts.UserPlatformVersion)
	opts.clusterVersion = versionDetect.Discovered
	if opts.clusterVersion == platform.MissingVersion {
		return fmt.Errorf("cannot autodetect the platform version, and no version given")
	}

	env.Log.Info("detection", "platform", opts.clusterPlatform, "reason", reason, "version", opts.clusterVersion, "source", source)
	if err := api.Deploy(env, api.Options{
		Platform: opts.clusterPlatform,
	}); err != nil {
		return err
	}
	if err := updaters.Deploy(env, commonOpts.UpdaterType, updaters.Options{
		Platform:        opts.clusterPlatform,
		PlatformVersion: opts.clusterVersion,
		WaitCompletion:  opts.waitCompletion,
		RTEConfigData:   commonOpts.RTEConfigData,
		DaemonSet:       daemonSetOptionsFromCommonOptions(commonOpts),
		EnableCRIHooks:  commonOpts.UpdaterCRIHooksEnable,
	}); err != nil {
		return err
	}
	if err := sched.Deploy(env, sched.Options{
		Platform:          opts.clusterPlatform,
		WaitCompletion:    opts.waitCompletion,
		Replicas:          int32(commonOpts.Replicas),
		RTEConfigData:     commonOpts.RTEConfigData,
		PullIfNotPresent:  commonOpts.PullIfNotPresent,
		CacheResyncPeriod: commonOpts.SchedResyncPeriod,
		CtrlPlaneAffinity: commonOpts.SchedCtrlPlaneAffinity,
		Verbose:           commonOpts.SchedVerbose,
	}); err != nil {
		return err
	}
	return nil
}
