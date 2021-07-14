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
	"github.com/fromanirh/deployer/pkg/deployer/sched"

	"github.com/fromanirh/deployer/pkg/clientutil"
	"github.com/fromanirh/deployer/pkg/deployer/rte"
	"github.com/fromanirh/deployer/pkg/manifests"

	"github.com/spf13/cobra"
)

type deployOptions struct{}

func NewDeployCommand(commonOpts *CommonOptions) *cobra.Command {
	opts := &deployOptions{}
	deploy := &cobra.Command{
		Use:   "deploy",
		Short: "deploy the components and configurations needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			return deployOnCluster(commonOpts, opts)
		},
		Args: cobra.NoArgs,
	}
	deploy.AddCommand(NewDeployAPICommand(commonOpts, opts))
	deploy.AddCommand(NewDeploySchedulerPluginCommand(commonOpts, opts))
	deploy.AddCommand(NewDeployTopologyUpdaterCommand(commonOpts, opts))
	return deploy
}

func NewRemoveCommand(commonOpts *CommonOptions) *cobra.Command {
	opts := &deployOptions{}
	remove := &cobra.Command{
		Use:   "remove",
		Short: "remove the components and configurations needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := Remove(opts); err != nil {
				return err
			}
			return rte.Remove(commonOpts.Log, rte.Options{})
		},
		Args: cobra.NoArgs,
	}
	remove.AddCommand(NewRemoveAPICommand(commonOpts, opts))
	remove.AddCommand(NewRemoveSchedulerPluginCommand(commonOpts, opts))
	remove.AddCommand(NewRemoveTopologyUpdaterCommand(commonOpts, opts))
	return remove
}

func NewDeployAPICommand(commonOpts *CommonOptions, opts *deployOptions) *cobra.Command {
	deploy := &cobra.Command{
		Use:   "api",
		Short: "deploy the APIs needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := Deploy(opts); err != nil {
				return err
			}
			return nil
		},
		Args: cobra.NoArgs,
	}
	return deploy
}

func NewDeploySchedulerPluginCommand(commonOpts *CommonOptions, opts *deployOptions) *cobra.Command {
	deploy := &cobra.Command{
		Use:   "scheduler-plugin",
		Short: "deploy the scheduler plugin needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			return sched.Deploy(commonOpts.Log, sched.Options{})
		},
		Args: cobra.NoArgs,
	}
	return deploy
}

func NewDeployTopologyUpdaterCommand(commonOpts *CommonOptions, opts *deployOptions) *cobra.Command {
	deploy := &cobra.Command{
		Use:   "topology-updater",
		Short: "deploy the topology updater needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			return rte.Deploy(commonOpts.Log, rte.Options{})
		},
		Args: cobra.NoArgs,
	}
	return deploy
}

func NewRemoveAPICommand(commonOpts *CommonOptions, opts *deployOptions) *cobra.Command {
	remove := &cobra.Command{
		Use:   "api",
		Short: "remove the APIs needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := Remove(opts); err != nil {
				return err
			}
			return nil
		},
		Args: cobra.NoArgs,
	}
	return remove
}

func NewRemoveSchedulerPluginCommand(commonOpts *CommonOptions, opts *deployOptions) *cobra.Command {
	remove := &cobra.Command{
		Use:   "scheduler-plugin",
		Short: "remove the scheduler plugin needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			return sched.Remove(commonOpts.Log, sched.Options{})
		},
		Args: cobra.NoArgs,
	}
	return remove
}

func NewRemoveTopologyUpdaterCommand(commonOpts *CommonOptions, opts *deployOptions) *cobra.Command {
	remove := &cobra.Command{
		Use:   "topology-updater",
		Short: "remove the topology updater needed for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			return rte.Remove(commonOpts.Log, rte.Options{})
		},
		Args: cobra.NoArgs,
	}
	return remove
}

func deployOnCluster(commonOpts *CommonOptions, opts *deployOptions) error {
	if err := Deploy(opts); err != nil {
		return err
	}
	return rte.Deploy(commonOpts.Log, rte.Options{})
}

func Deploy(opts *deployOptions) error {
	cs, err := clientutil.New()
	if err != nil {
		return err
	}

	crd, err := manifests.APICRD()
	if err != nil {
		return err
	}

	err = cs.Create(context.TODO(), crd)
	if err != nil {
		return err
	}
	return nil
}

func Remove(opts *deployOptions) error {
	cs, err := clientutil.New()
	if err != nil {
		return err
	}

	crd, err := manifests.APICRD()
	if err != nil {
		return err
	}

	err = cs.Delete(context.TODO(), crd)
	if err != nil {
		return err
	}
	return nil
}
