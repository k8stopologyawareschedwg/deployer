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

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/k8stopologyawareschedwg/deployer/pkg/commands"
	deployerversion "github.com/k8stopologyawareschedwg/deployer/pkg/version"
)

type versionOptions struct {
	fullOutput bool
	hashOnly   bool
}

func NewVersionCommand(commonOpts *commands.CommonOptions) *cobra.Command {
	opts := versionOptions{}
	version := &cobra.Command{
		Use:   "version",
		Short: "emit the version and exits succesfully",
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.hashOnly {
				fmt.Printf("%s\n", deployerversion.GitCommit)
			} else if opts.fullOutput {
				fmt.Printf("%s-%s\n", deployerversion.GitVersion, deployerversion.GitCommit[:9])
			} else {
				fmt.Printf("%s\n", deployerversion.GitVersion)
			}
			return nil
		},
		Args: cobra.NoArgs,
	}
	version.PersistentFlags().BoolVar(&opts.fullOutput, "full", false, "emit version and git hash.")
	version.PersistentFlags().BoolVar(&opts.hashOnly, "hash", false, "emit only the git hash.")
	return version
}

func main() {
	root := commands.NewRootCommand(NewVersionCommand)
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
