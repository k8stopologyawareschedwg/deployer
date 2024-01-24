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

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"github.com/k8stopologyawareschedwg/deployer/pkg/clientutil/nodes"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/options"
	"github.com/k8stopologyawareschedwg/deployer/pkg/validator"
)

type ValidateOutputMode int

const (
	ValidateOutputNone ValidateOutputMode = iota
	ValidateOutputText
	ValidateOutputJSON
	ValidateOutputLog
)

type validateOptions struct {
	outputMode ValidateOutputMode
	jsonOutput bool
}

func NewValidateCommand(env *deployer.Environment, commonOpts *options.Options) *cobra.Command {
	opts := &validateOptions{}
	validate := &cobra.Command{
		Use:   "validate",
		Short: "validate the cluster configuration to be correct for topology-aware-scheduling",
		RunE: func(cmd *cobra.Command, args []string) error {
			return validateCluster(cmd, env, commonOpts, opts, args)
		},
		Args: cobra.NoArgs,
	}
	validate.Flags().BoolVarP(&opts.jsonOutput, "json", "J", false, "output JSON, not text.")
	return validate
}

func validatePostSetupOptions(opts *validateOptions) error {
	if opts.outputMode != ValidateOutputNone {
		return nil // nothing to do!
	}
	opts.outputMode = ValidateOutputText
	if opts.jsonOutput {
		opts.outputMode = ValidateOutputJSON
	}
	return nil
}

type validationOutput struct {
	Success bool                         `json:"success"`
	Errors  []validator.ValidationResult `json:"errors,omitempty"`
}

func validateCluster(cmd *cobra.Command, env *deployer.Environment, commonOpts *options.Options, opts *validateOptions, args []string) error {
	// TODO
	validatePostSetupOptions(opts)

	err := env.EnsureClient()
	if err != nil {
		return err
	}

	vd, err := validator.NewValidator(env.Log)
	if err != nil {
		return err
	}

	nodeList, err := nodes.GetWorkers(env)
	if err != nil {
		return err
	}

	if _, err := vd.ValidateClusterConfig(nodeList); err != nil {
		return err
	}

	printValidationResults(vd.Results(), env.Log, opts.outputMode)
	return nil
}

// we need undecorated output, so we need to use fmt.Printf here. log packages add no value.
func printValidationResults(items []validator.ValidationResult, logger logr.Logger, outputMode ValidateOutputMode) {
	if len(items) == 0 {
		switch outputMode {
		case ValidateOutputJSON:
			json.NewEncoder(os.Stdout).Encode(validationOutput{
				Success: true,
			})
		case ValidateOutputText:
			fmt.Printf("PASSED>>: the cluster configuration looks ok!\n")
		case ValidateOutputLog:
			logger.Info("cluster configuration", "issue", "none")
		case ValidateOutputNone:
			fallthrough
		default:
			// do nothing!
		}
	} else {
		switch outputMode {
		case ValidateOutputJSON:
			json.NewEncoder(os.Stdout).Encode(validationOutput{
				Success: false,
				Errors:  items,
			})
		case ValidateOutputText:
			for idx, item := range items {
				fmt.Printf("ERROR#%03d: %s\n", idx, item.String())
			}
		case ValidateOutputLog:
			for idx, item := range items {
				logger.Info("cluster configuration", "issue", idx, "description", item.String())
			}
		case ValidateOutputNone:
			fallthrough
		default:
			// do nothing!
		}
	}
}
