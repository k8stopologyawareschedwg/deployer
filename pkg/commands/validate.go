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
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/k8stopologyawareschedwg/deployer/pkg/clientutil/nodes"
	"github.com/k8stopologyawareschedwg/deployer/pkg/validator"
)

type validateOptions struct {
	jsonOutput bool
}

func NewValidateCommand(commonOpts *CommonOptions) *cobra.Command {
	opts := &validateOptions{}
	validate := &cobra.Command{
		Use:   "validate",
		Short: "validate the cluster configuration to be correct for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			return validateCluster(cmd, commonOpts, opts, args)
		},
		Args: cobra.NoArgs,
	}
	validate.Flags().BoolVarP(&opts.jsonOutput, "json", "J", false, "output JSON, not text.")
	return validate
}

type validationOutput struct {
	Success bool                         `json:"success"`
	Errors  []validator.ValidationResult `json:"errors,omitempty"`
}

func validateCluster(cmd *cobra.Command, commonOpts *CommonOptions, opts *validateOptions, args []string) error {
	vd, err := validator.NewValidator(commonOpts.DebugLog)
	if err != nil {
		return err
	}

	nodeList, err := nodes.GetWorkers()
	if err != nil {
		return err
	}

	if _, err := vd.ValidateClusterConfig(nodeList); err != nil {
		return err
	}

	printValidationResults(vd.Results(), opts.jsonOutput)
	return nil
}

// we need undecorated output, so we need to use fmt.Printf here. log packages add no value.
func printValidationResults(items []validator.ValidationResult, jsonOutput bool) {
	if len(items) == 0 {
		if jsonOutput {
			json.NewEncoder(os.Stdout).Encode(validationOutput{
				Success: true,
			})
		} else {
			fmt.Printf("PASSED>>: the cluster configuration looks ok!\n")
		}
	} else {
		if jsonOutput {
			json.NewEncoder(os.Stdout).Encode(validationOutput{
				Success: false,
				Errors:  items,
			})
		} else {
			for idx, item := range items {
				fmt.Printf("ERROR#%03d: %s\n", idx, item.String())
			}
		}
	}
}
