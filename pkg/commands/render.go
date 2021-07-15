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
	"os"

	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/fromanirh/deployer/pkg/manifests"
	"github.com/fromanirh/deployer/pkg/manifests/api"
	"github.com/fromanirh/deployer/pkg/manifests/rte"
	"github.com/fromanirh/deployer/pkg/manifests/sched"
)

type renderOptions struct{}

func NewRenderCommand(commonOpts *CommonOptions) *cobra.Command {
	opts := &renderOptions{}
	render := &cobra.Command{
		Use:   "render",
		Short: "render all the manifests",
		RunE: func(cmd *cobra.Command, args []string) error {
			return renderManifests(cmd, commonOpts, opts, args)
		},
		Args: cobra.NoArgs,
	}
	return render
}

func renderManifests(cmd *cobra.Command, commonOpts *CommonOptions, opts *renderOptions, args []string) error {
	var objs []runtime.Object

	apiManifests, err := api.GetManifests()
	if err != nil {
		return err
	}
	objs = append(objs, apiManifests.UpdateNamespace().UpdatePullspecs().ToObjects()...)

	rteManifests, err := rte.GetManifests()
	if err != nil {
		return err
	}
	objs = append(objs, rteManifests.UpdateNamespace().UpdatePullspecs().ToObjects()...)

	schedManifests, err := sched.GetManifests()
	if err != nil {
		return err
	}
	objs = append(objs, schedManifests.UpdateNamespace().UpdatePullspecs().ToObjects()...)

	for _, obj := range objs {
		fmt.Printf("---\n")
		if err := manifests.SerializeObject(obj, os.Stdout); err != nil {
			return err
		}
	}

	return nil
}
