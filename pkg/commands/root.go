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
	Debug bool
	Log   *log.Logger
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
		Short: "WRITE ME",

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if commonOpts.Debug {
				commonOpts.Log = log.New(os.Stderr, "deployer ", log.LstdFlags)
			} else {
				commonOpts.Log = log.New(ioutil.Discard, "", 0)
			}
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
	)
	for _, extraCmd := range extraCmds {
		root.AddCommand(extraCmd(commonOpts))
	}

	return root
}
