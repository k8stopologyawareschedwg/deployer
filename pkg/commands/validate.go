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

	"github.com/fromanirh/deployer/pkg/clientutil/nodes"
	"github.com/fromanirh/deployer/pkg/validator"
)

type validateOptions struct{}

func NewValidateCommand(commonOpts *CommonOptions) *cobra.Command {
	opts := &validateOptions{}
	validate := &cobra.Command{
		Use:   "validate",
		Short: "validate the cluster configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return validateCluster(cmd, commonOpts, opts, args)
		},
		Args: cobra.NoArgs,
	}
	return validate
}

func validateCluster(cmd *cobra.Command, commonOpts *CommonOptions, opts *validateOptions, args []string) error {
	nodeList, err := nodes.GetWorkers()
	if err != nil {
		return err
	}

	vd := validator.Validator{
		Log: commonOpts.Log,
	}
	items, err := vd.ValidateClusterConfig(nodeList)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		fmt.Printf("PASSED>>: the cluster configuration looks ok!\n")
	} else {
		for idx, item := range items {
			fmt.Printf("ERROR#%03d: %s\n", idx, item.String())
		}
	}
	return nil
}
