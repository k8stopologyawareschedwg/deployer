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

	"github.com/fromanirh/deployer/pkg/manifests"
	"github.com/fromanirh/deployer/pkg/manifests/api"
	"github.com/fromanirh/deployer/pkg/manifests/rte"
	"github.com/fromanirh/deployer/pkg/manifests/sched"
	"github.com/fromanirh/deployer/pkg/tlog"
)

type renderOptions struct{}

func NewRenderCommand(commonOpts *CommonOptions) *cobra.Command {
	opts := &renderOptions{}
	render := &cobra.Command{
		Use:   "render",
		Short: "render all the manifests",
		RunE: func(cmd *cobra.Command, args []string) error {
			return renderManifests(cmd, commonOpts, opts, args)
		},
		Args: cobra.NoArgs,
	}
	render.AddCommand(NewRenderAPICommand(commonOpts, opts))
	render.AddCommand(NewRenderSchedulerPluginCommand(commonOpts, opts))
	render.AddCommand(NewRenderTopologyUpdaterCommand(commonOpts, opts))
	return render
}

func NewRenderAPICommand(commonOpts *CommonOptions, opts *renderOptions) *cobra.Command {
	render := &cobra.Command{
		Use:   "api",
		Short: "render the APIs needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiManifests, err := api.GetManifests(commonOpts.Platform)
			if err != nil {
				return err
			}
			return renderObjects(apiManifests.Update().ToObjects())
		},
		Args: cobra.NoArgs,
	}
	return render
}

func NewRenderSchedulerPluginCommand(commonOpts *CommonOptions, opts *renderOptions) *cobra.Command {
	render := &cobra.Command{
		Use:   "scheduler-plugin",
		Short: "render the scheduler plugin needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			schedManifests, err := sched.GetManifests(commonOpts.Platform)
			if err != nil {
				return err
			}
			rteManifests, err := rte.GetManifests(commonOpts.Platform)
			if err != nil {
				return fmt.Errorf("cannot get the rte manifests for sched: %w", err)
			}
			// no Options needed!
			rteManifests = rteManifests.Update(rte.UpdateOptions{})
			updateOpts := sched.UpdateOptions{
				Replicas:               int32(commonOpts.Replicas),
				NodeResourcesNamespace: rteManifests.DaemonSet.Namespace,
				PullIfNotPresent:       commonOpts.PullIfNotPresent,
			}
			la := tlog.NewLogAdapter(commonOpts.Log, commonOpts.DebugLog)
			return renderObjects(schedManifests.Update(la, updateOpts).ToObjects())
		},
		Args: cobra.NoArgs,
	}
	return render
}

func NewRenderTopologyUpdaterCommand(commonOpts *CommonOptions, opts *renderOptions) *cobra.Command {
	render := &cobra.Command{
		Use:   "topology-updater",
		Short: "render the topology updater needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			rteManifests, err := rte.GetManifests(commonOpts.Platform)
			if err != nil {
				return err
			}
			updateOpts := rte.UpdateOptions{
				ConfigData:       commonOpts.RTEConfigData,
				PullIfNotPresent: commonOpts.PullIfNotPresent,
			}
			return renderObjects(rteManifests.Update(updateOpts).ToObjects())
		},
		Args: cobra.NoArgs,
	}
	return render
}

func renderManifests(cmd *cobra.Command, commonOpts *CommonOptions, opts *renderOptions, args []string) error {
	var objs []client.Object

	apiManifests, err := api.GetManifests(commonOpts.Platform)
	if err != nil {
		return err
	}
	objs = append(objs, apiManifests.Update().ToObjects()...)

	rteManifests, err := rte.GetManifests(commonOpts.Platform)
	if err != nil {
		return err
	}
	rteUpdateOpts := rte.UpdateOptions{
		ConfigData:       commonOpts.RTEConfigData,
		PullIfNotPresent: commonOpts.PullIfNotPresent,
	}
	objs = append(objs, rteManifests.Update(rteUpdateOpts).ToObjects()...)

	schedManifests, err := sched.GetManifests(commonOpts.Platform)
	if err != nil {
		return err
	}

	schedUpdateOpts := sched.UpdateOptions{
		Replicas:               int32(commonOpts.Replicas),
		NodeResourcesNamespace: rteManifests.Namespace.Name,
		PullIfNotPresent:       commonOpts.PullIfNotPresent,
	}

	la := tlog.NewLogAdapter(commonOpts.Log, commonOpts.DebugLog)
	objs = append(objs, schedManifests.Update(la, schedUpdateOpts).ToObjects()...)

	return renderObjects(objs)
}

func renderObjects(objs []client.Object) error {
	for _, obj := range objs {
		fmt.Printf("---\n")
		if err := manifests.SerializeObject(obj, os.Stdout); err != nil {
			return err
		}
	}

	return nil
}
