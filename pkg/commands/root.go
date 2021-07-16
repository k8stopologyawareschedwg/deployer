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
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/cobra"
)

type CommonOptions struct {
	Debug    bool
	Log      *log.Logger
	DebugLog *log.Logger
}

func ShowHelp(cmd *cobra.Command, args []string) error {
	fmt.Fprint(cmd.OutOrStderr(), cmd.UsageString())
	return nil
}

type NewCommandFunc func(ko *CommonOptions) *cobra.Command

// NewRootCommand returns entrypoint command to interact with all other commands
func NewRootCommand(extraCmds ...NewCommandFunc) *cobra.Command {
	commonOpts := &CommonOptions{}

	root := &cobra.Command{
		Use:   "deployer",
		Short: "deployer helps setting up all the topology-aware-scheduling components on a kubernetes cluster",

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if commonOpts.Debug {
				commonOpts.DebugLog = log.New(os.Stderr, "", log.LstdFlags)
			} else {
				commonOpts.DebugLog = log.New(ioutil.Discard, "", 0)
			}
			// we abuse the logger to have a common interface and the timestamps
			commonOpts.Log = log.New(os.Stdout, "", log.LstdFlags)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return ShowHelp(cmd, args)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().BoolVarP(&commonOpts.Debug, "debug", "D", false, "enable debug log")

	root.AddCommand(
		NewRenderCommand(commonOpts),
		NewValidateCommand(commonOpts),
		NewDeployCommand(commonOpts),
		NewRemoveCommand(commonOpts),
		NewSetupCommand(commonOpts),
	)
	for _, extraCmd := range extraCmds {
		root.AddCommand(extraCmd(commonOpts))
	}

	return root
}
