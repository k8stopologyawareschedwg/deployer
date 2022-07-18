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

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform/detect"
)

type detectOptions struct {
	jsonOutput bool
}

func NewDetectCommand(commonOpts *CommonOptions) *cobra.Command {
	opts := &detectOptions{}
	detect := &cobra.Command{
		Use:   "detect",
		Short: "detect the cluster platform (kubernetes, openshift...)",
		RunE: func(cmd *cobra.Command, args []string) error {
			platKind, reason, _ := detect.FindPlatform(commonOpts.UserPlatform)
			commonOpts.DebugLog.Printf("platform kind %s (%s)", platKind.Discovered, reason)
			platVer, reason, _ := detect.FindVersion(platKind.Discovered, commonOpts.UserPlatformVersion)
			commonOpts.DebugLog.Printf("platform version %s (%s)", platVer.Discovered, reason)

			cluster := detect.ClusterInfo{
				Platform: platKind,
				Version:  platVer,
			}
			if opts.jsonOutput {
				json.NewEncoder(os.Stdout).Encode(cluster)
			} else {
				fmt.Printf("%s:%s\n", cluster.Platform.Discovered, cluster.Version.Discovered)
			}
			return nil
		},
		Args: cobra.NoArgs,
	}
	detect.Flags().BoolVarP(&opts.jsonOutput, "json", "J", false, "output JSON, not text.")
	return detect
}
