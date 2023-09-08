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
	"os"

	"github.com/spf13/cobra"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deploy"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/updaters"
	"github.com/k8stopologyawareschedwg/deployer/pkg/images"
)

type ImagesOptions struct {
	jsonOutput bool
	rawOutput  bool
	useSHA     bool
}

func NewImagesCommand(env *deployer.Environment, commonOpts *deploy.Options) *cobra.Command {
	opts := &ImagesOptions{}
	images := &cobra.Command{
		Use:   "images",
		Short: "dump the container images used to deploy",
		RunE: func(cmd *cobra.Command, args []string) error {
			images.SetDefaults(opts.useSHA)
			updaterImage := getUpdaterImage(commonOpts.UpdaterType)
			fk := images.FormatText
			if opts.jsonOutput {
				fk = images.FormatJSON
			}
			imo := images.NewOutput(updaterImage)
			var of images.Formatter = imo
			if opts.rawOutput {
				of = imo.ToList()
			}
			of.Format(fk, os.Stdout)
			return nil
		},
		Args: cobra.NoArgs,
	}
	images.Flags().BoolVarP(&opts.jsonOutput, "json", "J", false, "output JSON, not text (default).")
	images.Flags().BoolVarP(&opts.rawOutput, "raw", "r", false, "output raw list. Default is key=value object.")
	images.Flags().BoolVarP(&opts.useSHA, "sha", "S", false, "emit SHA256 pullspects, not tag pullspecs.")
	return images
}

func getUpdaterImage(updaterType string) string {
	if updaterType == updaters.RTE {
		return images.ResourceTopologyExporterImage
	}
	return images.NodeFeatureDiscoveryImage
}
