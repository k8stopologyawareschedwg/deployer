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
	"context"
	"fmt"

	"github.com/spf13/cobra"

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

func NewDeployCommand(commonOpts *CommonOptions) *cobra.Command {
	opts := &DeployOptions{}
	deploy := &cobra.Command{
		Use:   "deploy",
		Short: "deploy the components and configurations needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			return deployOnCluster(commonOpts, opts)
		},
		Args: cobra.NoArgs,
	}
	deploy.PersistentFlags().BoolVarP(&opts.waitCompletion, "wait", "W", false, "wait for deployment to be all completed.")
	deploy.AddCommand(NewDeployAPICommand(commonOpts, opts))
	deploy.AddCommand(NewDeploySchedulerPluginCommand(commonOpts, opts))
	deploy.AddCommand(NewDeployTopologyUpdaterCommand(commonOpts, opts))
	return deploy
}

func NewRemoveCommand(commonOpts *CommonOptions) *cobra.Command {
	opts := &DeployOptions{}
	remove := &cobra.Command{
		Use:   "remove",
		Short: "remove the components and configurations needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			env, err := environFromOpts(commonOpts)
			if err != nil {
				return err
			}

			ctx := context.Background()

			platDetect, reason, _ := detect.FindPlatform(ctx, commonOpts.UserPlatform)
			opts.clusterPlatform = platDetect.Discovered
			if opts.clusterPlatform == platform.Unknown {
				return fmt.Errorf("cannot autodetect the platform, and no platform given")
			}
			versionDetect, source, _ := detect.FindVersion(ctx, platDetect.Discovered, commonOpts.UserPlatformVersion)
			opts.clusterVersion = versionDetect.Discovered
			if opts.clusterVersion == platform.MissingVersion {
				return fmt.Errorf("cannot autodetect the platform version, and no version given")
			}
			commonOpts.DebugLog.Info("detection", "platform", opts.clusterPlatform, "reason", reason, "version", opts.clusterVersion, "source", source)

			err = sched.Remove(env, sched.Options{
				Platform:          opts.clusterPlatform,
				WaitCompletion:    opts.waitCompletion,
				Replicas:          int32(commonOpts.Replicas),
				RTEConfigData:     commonOpts.RTEConfigData,
				PullIfNotPresent:  commonOpts.PullIfNotPresent,
				CacheResyncPeriod: commonOpts.schedResyncPeriod,
			})
			if err != nil {
				// intentionally keep going to remove as much as possible
				commonOpts.Log.Info("while removing", "error", err)
			}
			err = updaters.Remove(env, commonOpts.UpdaterType, updaters.Options{
				Platform:         opts.clusterPlatform,
				PlatformVersion:  opts.clusterVersion,
				WaitCompletion:   opts.waitCompletion,
				PullIfNotPresent: commonOpts.PullIfNotPresent,
				RTEConfigData:    commonOpts.RTEConfigData,
			})
			if err != nil {
				// intentionally keep going to remove as much as possible
				commonOpts.Log.Info("while removing", "error", err)
			}
			err = api.Remove(env, api.Options{
				Platform: opts.clusterPlatform,
			})
			if err != nil {
				// intentionally keep going to remove as much as possible
				commonOpts.Log.Info("while removing", "error", err)
			}
			return nil
		},
		Args: cobra.NoArgs,
	}
	remove.PersistentFlags().BoolVarP(&opts.waitCompletion, "wait", "W", false, "wait for removal to be all completed.")
	remove.AddCommand(NewRemoveAPICommand(commonOpts, opts))
	remove.AddCommand(NewRemoveSchedulerPluginCommand(commonOpts, opts))
	remove.AddCommand(NewRemoveTopologyUpdaterCommand(commonOpts, opts))
	return remove
}

func NewDeployAPICommand(commonOpts *CommonOptions, opts *DeployOptions) *cobra.Command {
	deploy := &cobra.Command{
		Use:   "api",
		Short: "deploy the APIs needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			env, err := environFromOpts(commonOpts)
			if err != nil {
				return err
			}

			ctx := context.Background()

			platDetect, reason, _ := detect.FindPlatform(ctx, commonOpts.UserPlatform)
			opts.clusterPlatform = platDetect.Discovered
			if opts.clusterPlatform == platform.Unknown {
				return fmt.Errorf("cannot autodetect the platform, and no platform given")
			}
			versionDetect, source, _ := detect.FindVersion(ctx, platDetect.Discovered, commonOpts.UserPlatformVersion)
			opts.clusterVersion = versionDetect.Discovered
			if opts.clusterVersion == platform.MissingVersion {
				return fmt.Errorf("cannot autodetect the platform version, and no version given")
			}

			commonOpts.DebugLog.Info("detection", "platform", opts.clusterPlatform, "reason", reason, "version", opts.clusterVersion, "source", source)
			if err := api.Deploy(env, api.Options{Platform: opts.clusterPlatform}); err != nil {
				return err
			}
			return nil
		},
		Args: cobra.NoArgs,
	}
	return deploy
}

func NewDeploySchedulerPluginCommand(commonOpts *CommonOptions, opts *DeployOptions) *cobra.Command {
	deploy := &cobra.Command{
		Use:   "scheduler-plugin",
		Short: "deploy the scheduler plugin needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			env, err := environFromOpts(commonOpts)
			if err != nil {
				return err
			}

			ctx := context.Background()

			platDetect, reason, _ := detect.FindPlatform(ctx, commonOpts.UserPlatform)
			opts.clusterPlatform = platDetect.Discovered
			if opts.clusterPlatform == platform.Unknown {
				return fmt.Errorf("cannot autodetect the platform, and no platform given")
			}
			versionDetect, source, _ := detect.FindVersion(ctx, platDetect.Discovered, commonOpts.UserPlatformVersion)
			opts.clusterVersion = versionDetect.Discovered
			if opts.clusterVersion == platform.MissingVersion {
				return fmt.Errorf("cannot autodetect the platform version, and no version given")
			}

			commonOpts.DebugLog.Info("detection", "platform", opts.clusterPlatform, "reason", reason, "version", opts.clusterVersion, "source", source)
			return sched.Deploy(env, sched.Options{
				Platform:          opts.clusterPlatform,
				WaitCompletion:    opts.waitCompletion,
				Replicas:          int32(commonOpts.Replicas),
				RTEConfigData:     commonOpts.RTEConfigData,
				PullIfNotPresent:  commonOpts.PullIfNotPresent,
				CacheResyncPeriod: commonOpts.schedResyncPeriod,
			})
		},
		Args: cobra.NoArgs,
	}
	return deploy
}

func NewDeployTopologyUpdaterCommand(commonOpts *CommonOptions, opts *DeployOptions) *cobra.Command {
	deploy := &cobra.Command{
		Use:   "topology-updater",
		Short: "deploy the topology updater needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			env, err := environFromOpts(commonOpts)
			if err != nil {
				return err
			}

			ctx := context.Background()

			platDetect, reason, _ := detect.FindPlatform(ctx, commonOpts.UserPlatform)
			opts.clusterPlatform = platDetect.Discovered
			if opts.clusterPlatform == platform.Unknown {
				return fmt.Errorf("cannot autodetect the platform, and no platform given")
			}
			versionDetect, source, _ := detect.FindVersion(ctx, platDetect.Discovered, commonOpts.UserPlatformVersion)
			opts.clusterVersion = versionDetect.Discovered
			if opts.clusterVersion == platform.MissingVersion {
				return fmt.Errorf("cannot autodetect the platform version, and no version given")
			}

			commonOpts.DebugLog.Info("detection", "platform", opts.clusterPlatform, "reason", reason, "version", opts.clusterVersion, "source", source)
			return updaters.Deploy(env, commonOpts.UpdaterType, updaters.Options{
				Platform:         opts.clusterPlatform,
				PlatformVersion:  opts.clusterVersion,
				WaitCompletion:   opts.waitCompletion,
				PullIfNotPresent: commonOpts.PullIfNotPresent,
				RTEConfigData:    commonOpts.RTEConfigData,
			})
		},
		Args: cobra.NoArgs,
	}
	return deploy
}

func NewRemoveAPICommand(commonOpts *CommonOptions, opts *DeployOptions) *cobra.Command {
	remove := &cobra.Command{
		Use:   "api",
		Short: "remove the APIs needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			env, err := environFromOpts(commonOpts)
			if err != nil {
				return err
			}

			ctx := context.Background()

			platDetect, reason, _ := detect.FindPlatform(ctx, commonOpts.UserPlatform)
			opts.clusterPlatform = platDetect.Discovered
			if opts.clusterPlatform == platform.Unknown {
				return fmt.Errorf("cannot autodetect the platform, and no platform given")
			}
			versionDetect, source, _ := detect.FindVersion(ctx, platDetect.Discovered, commonOpts.UserPlatformVersion)
			opts.clusterVersion = versionDetect.Discovered
			if opts.clusterVersion == platform.MissingVersion {
				return fmt.Errorf("cannot autodetect the platform version, and no version given")
			}

			commonOpts.DebugLog.Info("detection", "platform", opts.clusterPlatform, "reason", reason, "version", opts.clusterVersion, "source", source)
			if err := api.Remove(env, api.Options{Platform: opts.clusterPlatform}); err != nil {
				return err
			}
			return nil
		},
		Args: cobra.NoArgs,
	}
	return remove
}

func NewRemoveSchedulerPluginCommand(commonOpts *CommonOptions, opts *DeployOptions) *cobra.Command {
	remove := &cobra.Command{
		Use:   "scheduler-plugin",
		Short: "remove the scheduler plugin needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			env, err := environFromOpts(commonOpts)
			if err != nil {
				return err
			}

			ctx := context.Background()

			platDetect, reason, _ := detect.FindPlatform(ctx, commonOpts.UserPlatform)
			opts.clusterPlatform = platDetect.Discovered
			if opts.clusterPlatform == platform.Unknown {
				return fmt.Errorf("cannot autodetect the platform, and no platform given")
			}
			versionDetect, source, _ := detect.FindVersion(ctx, platDetect.Discovered, commonOpts.UserPlatformVersion)
			opts.clusterVersion = versionDetect.Discovered
			if opts.clusterVersion == platform.MissingVersion {
				return fmt.Errorf("cannot autodetect the platform version, and no version given")
			}

			commonOpts.DebugLog.Info("detection", "platform", opts.clusterPlatform, "reason", reason, "version", opts.clusterVersion, "source", source)
			return sched.Remove(env, sched.Options{
				Platform:          opts.clusterPlatform,
				WaitCompletion:    opts.waitCompletion,
				Replicas:          int32(commonOpts.Replicas),
				RTEConfigData:     commonOpts.RTEConfigData,
				PullIfNotPresent:  commonOpts.PullIfNotPresent,
				CacheResyncPeriod: commonOpts.schedResyncPeriod,
			})
		},
		Args: cobra.NoArgs,
	}
	return remove
}

func NewRemoveTopologyUpdaterCommand(commonOpts *CommonOptions, opts *DeployOptions) *cobra.Command {
	remove := &cobra.Command{
		Use:   "topology-updater",
		Short: "remove the topology updater needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			env, err := environFromOpts(commonOpts)
			if err != nil {
				return err
			}

			ctx := context.Background()

			platDetect, reason, _ := detect.FindPlatform(ctx, commonOpts.UserPlatform)
			opts.clusterPlatform = platDetect.Discovered
			if opts.clusterPlatform == platform.Unknown {
				return fmt.Errorf("cannot autodetect the platform, and no platform given")
			}
			versionDetect, source, _ := detect.FindVersion(ctx, platDetect.Discovered, commonOpts.UserPlatformVersion)
			opts.clusterVersion = versionDetect.Discovered
			if opts.clusterVersion == platform.MissingVersion {
				return fmt.Errorf("cannot autodetect the platform version, and no version given")
			}

			commonOpts.DebugLog.Info("detection", "platform", opts.clusterPlatform, "reason", reason, "version", opts.clusterVersion, "source", source)
			return updaters.Remove(env, commonOpts.UpdaterType, updaters.Options{
				Platform:         opts.clusterPlatform,
				PlatformVersion:  opts.clusterVersion,
				WaitCompletion:   opts.waitCompletion,
				PullIfNotPresent: commonOpts.PullIfNotPresent,
				RTEConfigData:    commonOpts.RTEConfigData,
			})
		},
		Args: cobra.NoArgs,
	}
	return remove
}

func deployOnCluster(commonOpts *CommonOptions, opts *DeployOptions) error {
	env, err := environFromOpts(commonOpts)
	if err != nil {
		return err
	}

	ctx := context.Background()

	platDetect, reason, _ := detect.FindPlatform(ctx, commonOpts.UserPlatform)
	opts.clusterPlatform = platDetect.Discovered
	if opts.clusterPlatform == platform.Unknown {
		return fmt.Errorf("cannot autodetect the platform, and no platform given")
	}
	versionDetect, source, _ := detect.FindVersion(ctx, platDetect.Discovered, commonOpts.UserPlatformVersion)
	opts.clusterVersion = versionDetect.Discovered
	if opts.clusterVersion == platform.MissingVersion {
		return fmt.Errorf("cannot autodetect the platform version, and no version given")
	}

	commonOpts.DebugLog.Info("detection", "platform", opts.clusterPlatform, "reason", reason, "version", opts.clusterVersion, "source", source)
	if err := api.Deploy(env, api.Options{
		Platform: opts.clusterPlatform,
	}); err != nil {
		return err
	}
	if err := updaters.Deploy(env, commonOpts.UpdaterType, updaters.Options{
		Platform:         opts.clusterPlatform,
		PlatformVersion:  opts.clusterVersion,
		WaitCompletion:   opts.waitCompletion,
		PullIfNotPresent: commonOpts.PullIfNotPresent,
		RTEConfigData:    commonOpts.RTEConfigData,
	}); err != nil {
		return err
	}
	if err := sched.Deploy(env, sched.Options{
		Platform:          opts.clusterPlatform,
		WaitCompletion:    opts.waitCompletion,
		Replicas:          int32(commonOpts.Replicas),
		RTEConfigData:     commonOpts.RTEConfigData,
		PullIfNotPresent:  commonOpts.PullIfNotPresent,
		CacheResyncPeriod: commonOpts.schedResyncPeriod,
	}); err != nil {
		return err
	}
	return nil
}
