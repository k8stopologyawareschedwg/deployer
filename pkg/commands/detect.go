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

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform/detect"
	"github.com/k8stopologyawareschedwg/deployer/pkg/options"
)

type detectOptions struct {
	controlPlane bool
	jsonOutput   bool
}

func NewDetectCommand(env *deployer.Environment, commonOpts *options.Options) *cobra.Command {
	opts := &detectOptions{}
	detect := &cobra.Command{
		Use:   "detect",
		Short: "detect the cluster platform (kubernetes, openshift...)",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			if err = env.EnsureClient(); err != nil {
				return err
			}

			if opts.controlPlane {
				info, _ := detect.ControlPlaneFromLister(env.Ctx, env.Cli)
				serialize(opts, info)
				return nil

			}

			platKind, kindReason, _ := detect.FindPlatform(env.Ctx, commonOpts.UserPlatform)
			platVer, verReason, _ := detect.FindVersion(env.Ctx, platKind.Discovered, commonOpts.UserPlatformVersion)

			env.Log.V(3).Info("detection", "platform", platKind, "reason", kindReason, "version", platVer, "source", verReason)

			cluster := detect.ClusterInfo{
				Platform: platKind,
				Version:  platVer,
			}
			serialize(opts, cluster)
			return nil
		},
		Args: cobra.NoArgs,
	}
	detect.Flags().BoolVarP(&opts.jsonOutput, "json", "J", false, "output JSON, not text.")
	detect.Flags().BoolVar(&opts.controlPlane, "control-plane", false, "detect control plane info, not cluster info")
	return detect
}

type serializer interface {
	ToJSON() string
	String() string
}

func serialize(opts *detectOptions, sr serializer) {
	var out string
	if opts.jsonOutput {
		out = sr.ToJSON()
	} else {
		out = sr.String()
	}
	fmt.Printf("%s\n", out)
}
