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
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/k8stopologyawareschedwg/deployer/pkg/images"
)

type imagesOptions struct {
	jsonOutput bool
	rawOutput  bool
}

func NewImagesCommand(commonOpts *CommonOptions) *cobra.Command {
	opts := &imagesOptions{}
	images := &cobra.Command{
		Use:   "images",
		Short: "dump the container images used to deploy",
		RunE: func(cmd *cobra.Command, args []string) error {
			imo := newImageOutput()
			if opts.rawOutput {
				il := imo.ToList()
				if opts.jsonOutput {
					il.EncodeJSON(os.Stdout)
				} else {
					il.EncodeText(os.Stdout)
				}
			} else {
				if opts.jsonOutput {
					imo.EncodeJSON(os.Stdout)
				} else {
					imo.EncodeText(os.Stdout)
				}
			}
			return nil
		},
		Args: cobra.NoArgs,
	}
	images.Flags().BoolVarP(&opts.jsonOutput, "json", "J", false, "output JSON, not text (default).")
	images.Flags().BoolVarP(&opts.rawOutput, "raw", "r", false, "output raw list. Default is key=value object.")
	return images
}

type imageOutput struct {
	TopologyUpdater     string `json:"topology_updater"`
	SchedulerPlugin     string `json:"scheduler_plugin"`
	SchedulerController string `json:"scheduler_controller"`
}

func newImageOutput() imageOutput {
	return imageOutput{
		TopologyUpdater:     images.ResourceTopologyExporterDefaultImageTag,
		SchedulerPlugin:     images.SchedulerPluginSchedulerDefaultImageTag,
		SchedulerController: images.SchedulerPluginControllerDefaultImageTag,
	}
}

type imageList []string

func (imo imageOutput) ToList() imageList {
	return []string{
		imo.TopologyUpdater,
		imo.SchedulerPlugin,
		imo.SchedulerController,
	}
}

func (il imageList) EncodeText(w io.Writer) {
	fmt.Fprintf(w, "%s\n", strings.Join(il, "\n"))
}

func (il imageList) EncodeJSON(w io.Writer) {
	json.NewEncoder(os.Stdout).Encode(il)
}

func (imo imageOutput) EncodeText(w io.Writer) {
	fmt.Fprintf(w, "TAS_SCHEDULER_PLUGIN_IMAGE=%s\n", imo.SchedulerPlugin)
	fmt.Fprintf(w, "TAS_SCHEDULER_PLUGIN_CONTROLLER_IMAGE=%s\n", imo.SchedulerController)
	fmt.Fprintf(w, "TAS_RESOURCE_EXPORTER_IMAGE=%s\n", imo.TopologyUpdater)
}

func (imo imageOutput) EncodeJSON(w io.Writer) {
	json.NewEncoder(w).Encode(imo)
}
