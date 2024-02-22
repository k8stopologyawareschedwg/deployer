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
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/controller-runtime/pkg/client"

	selinuxassets "github.com/k8stopologyawareschedwg/deployer/pkg/assets/selinux"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/updaters"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests/api"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests/sched"
	"github.com/k8stopologyawareschedwg/deployer/pkg/options"
)

func NewRenderCommand(env *deployer.Environment, commonOpts *options.Options) *cobra.Command {
	opts := &options.Scheduler{}
	render := &cobra.Command{
		Use:   "render",
		Short: "render all the manifests",
		RunE: func(cmd *cobra.Command, args []string) error {
			if commonOpts.UserPlatform == platform.Unknown {
				return fmt.Errorf("must explicitly select a cluster platform")
			}
			return RenderManifests(env, commonOpts)
		},
		Args: cobra.NoArgs,
	}
	render.AddCommand(NewRenderAPICommand(env, commonOpts, opts))
	render.AddCommand(NewRenderSchedulerPluginCommand(env, commonOpts, opts))
	render.AddCommand(NewRenderTopologyUpdaterCommand(env, commonOpts, opts))
	render.AddCommand(NewRenderPolicyCommand(env, commonOpts, opts))
	return render
}

func NewRenderAPICommand(env *deployer.Environment, commonOpts *options.Options, opts *options.Scheduler) *cobra.Command {
	render := &cobra.Command{
		Use:   "api",
		Short: "render the APIs needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			if commonOpts.UserPlatform == platform.Unknown {
				return fmt.Errorf("must explicitly select a cluster platform")
			}
			apiManifests, err := api.GetManifests(commonOpts.UserPlatform)
			if err != nil {
				return err
			}
			apiObjs, err := apiManifests.Render()
			if err != nil {
				return err
			}
			return manifests.RenderObjects(apiObjs.ToObjects(), os.Stdout)
		},
		Args: cobra.NoArgs,
	}
	return render
}

func NewRenderSchedulerPluginCommand(env *deployer.Environment, commonOpts *options.Options, opts *options.Scheduler) *cobra.Command {
	render := &cobra.Command{
		Use:   "scheduler-plugin",
		Short: "render the scheduler plugin needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			if commonOpts.UserPlatform == platform.Unknown {
				return fmt.Errorf("must explicitly select a cluster platform")
			}

			_, namespace, err := updaters.SetupNamespace(commonOpts.UpdaterType)
			if err != nil {
				return err
			}

			schedManifests, err := sched.GetManifests(commonOpts.UserPlatform, namespace)
			if err != nil {
				return err
			}

			renderOpts := options.Scheduler{
				Replicas:               int32(commonOpts.Replicas),
				PullIfNotPresent:       commonOpts.PullIfNotPresent,
				ScoringStratConfigData: commonOpts.SchedScoringStratConfigData,
				CacheParamsConfigData:  commonOpts.SchedCacheParamsConfigData,
			}
			schedObjs, err := schedManifests.Render(env.Log, renderOpts)
			if err != nil {
				return err
			}
			return manifests.RenderObjects(schedObjs.ToObjects(), os.Stdout)
		},
		Args: cobra.NoArgs,
	}
	return render
}

func NewRenderTopologyUpdaterCommand(env *deployer.Environment, commonOpts *options.Options, opts *options.Scheduler) *cobra.Command {
	render := &cobra.Command{
		Use:   "topology-updater",
		Short: "render the topology updater needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			if commonOpts.UserPlatform == platform.Unknown {
				return fmt.Errorf("must explicitly select a cluster platform")
			}
			objs, _, err := makeUpdaterObjects(commonOpts)
			if err != nil {
				return err
			}
			return manifests.RenderObjects(objs, os.Stdout)
		},
		Args: cobra.NoArgs,
	}
	return render
}

func makeUpdaterObjects(commonOpts *options.Options) ([]client.Object, string, error) {
	ns, namespace, err := updaters.SetupNamespace(commonOpts.UpdaterType)
	if err != nil {
		return nil, namespace, err
	}

	opts := options.Updater{
		PlatformVersion: commonOpts.UserPlatformVersion,
		Platform:        commonOpts.UserPlatform,
		RTEConfigData:   commonOpts.RTEConfigData,
		DaemonSet:       options.ForDaemonSet(commonOpts),
		EnableCRIHooks:  commonOpts.UpdaterCRIHooksEnable,
	}
	objs, err := updaters.GetObjects(opts, commonOpts.UpdaterType, namespace)
	if err != nil {
		return nil, namespace, err
	}

	return append([]client.Object{ns}, objs...), namespace, nil
}

func RenderManifests(env *deployer.Environment, commonOpts *options.Options) error {
	var objs []client.Object

	apiManifests, err := api.GetManifests(commonOpts.UserPlatform)
	if err != nil {
		return err
	}
	apiObjs, err := apiManifests.Render()
	if err != nil {
		return err
	}
	objs = append(objs, apiObjs.ToObjects()...)

	updaterObjs, updaterNs, err := makeUpdaterObjects(commonOpts)
	if err != nil {
		return err
	}
	objs = append(objs, updaterObjs...)

	schedManifests, err := sched.GetManifests(commonOpts.UserPlatform, updaterNs)
	if err != nil {
		return err
	}

	schedRenderOpts := options.Scheduler{
		Replicas:               int32(commonOpts.Replicas),
		PullIfNotPresent:       commonOpts.PullIfNotPresent,
		ProfileName:            commonOpts.SchedProfileName,
		CacheResyncPeriod:      commonOpts.SchedResyncPeriod,
		CtrlPlaneAffinity:      commonOpts.SchedCtrlPlaneAffinity,
		Verbose:                commonOpts.SchedVerbose,
		ScoringStratConfigData: commonOpts.SchedScoringStratConfigData,
		CacheParamsConfigData:  commonOpts.SchedCacheParamsConfigData,
	}

	schedObjs, err := schedManifests.Render(env.Log, schedRenderOpts)
	if err != nil {
		return err
	}
	objs = append(objs, schedObjs.ToObjects()...)

	return manifests.RenderObjects(objs, os.Stdout)
}

func NewRenderPolicyCommand(env *deployer.Environment, commonOpts *options.Options, opts *options.Scheduler) *cobra.Command {
	render := &cobra.Command{
		Use:   "policy",
		Short: "render the SELinux policy needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			if commonOpts.UserPlatform != platform.OpenShift {
				return fmt.Errorf("must explicitly select the OpenShift platform")
			}
			selinuxPolicy, err := selinuxassets.GetPolicy(commonOpts.UserPlatformVersion)
			if err != nil {
				return err
			}
			_, err = os.Stdout.Write(selinuxPolicy)
			return err
		},
		Args: cobra.NoArgs,
	}
	return render
}
