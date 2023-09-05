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
	"github.com/k8stopologyawareschedwg/deployer/pkg/deploy"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/spf13/cobra"
)

func NewSetupCommand(env *deployer.Environment, commonOpts *deploy.Options) *cobra.Command {
	depOpts := &DeployOptions{}
	valOpts := &validateOptions{
		outputMode: ValidateOutputLog,
	}
	setup := &cobra.Command{
		Use:   "setup",
		Short: "validate and setup a cluster to be used for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateCluster(cmd, env, commonOpts, valOpts, args); err != nil {
				return err
			}
			return deployOnCluster(env, commonOpts, depOpts)
		},
		Args: cobra.NoArgs,
	}
	return setup
}
