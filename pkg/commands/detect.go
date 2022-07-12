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
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
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
			platKind := detectPlatform(commonOpts.DebugLog, commonOpts.UserPlatform)
			platVer := detectVersion(commonOpts.DebugLog, platKind.Discovered, commonOpts.UserPlatformVersion)
			cluster := clusterDetection{
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

type platformDetection struct {
	AutoDetected platform.Platform `json:"auto_detected"`
	UserSupplied platform.Platform `json:"user_supplied"`
	Discovered   platform.Platform `json:"discovered"`
}

type versionDetection struct {
	AutoDetected platform.Version `json:"auto_detected"`
	UserSupplied platform.Version `json:"user_supplied"`
	Discovered   platform.Version `json:"discovered"`
}

type clusterDetection struct {
	Platform platformDetection `json:"platform"`
	Version  versionDetection  `json:"version"`
}

func detectPlatform(debugLog *log.Logger, userSupplied platform.Platform) platformDetection {
	do := platformDetection{
		AutoDetected: platform.Unknown,
		UserSupplied: userSupplied,
		Discovered:   platform.Unknown,
	}

	if do.UserSupplied != platform.Unknown {
		debugLog.Printf("user-supplied platform: %q", do.UserSupplied)
		do.Discovered = do.UserSupplied
		return do
	}

	dp, err := detect.Platform()
	if err != nil {
		debugLog.Printf("failed to detect the platform: %v", err)
		return do
	}

	debugLog.Printf("auto-detected platform: %q", dp)
	do.AutoDetected = dp
	do.Discovered = do.AutoDetected
	return do
}

func detectVersion(debugLog *log.Logger, plat platform.Platform, userSupplied platform.Version) versionDetection {
	do := versionDetection{
		AutoDetected: platform.MissingVersion,
		UserSupplied: userSupplied,
		Discovered:   platform.MissingVersion,
	}

	if do.UserSupplied != platform.MissingVersion {
		debugLog.Printf("user-supplied version: %q", do.UserSupplied)
		do.Discovered = do.UserSupplied
		return do
	}

	dv, err := detect.Version(plat)
	if err != nil {
		debugLog.Printf("failed to detect the version: %v", err)
		return do
	}

	debugLog.Printf("auto-detected version: %q", dv)
	do.AutoDetected = dv
	do.Discovered = do.AutoDetected
	return do
}
