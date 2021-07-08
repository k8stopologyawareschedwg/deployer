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

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/fromanirh/deployer/pkg/manifests"
)

type renderOptions struct {
	pidIdent string
}

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

	crd, err := manifests.LoadAPICRD()
	if err != nil {
		return err
	}
	objs = append(objs, crd)

	rteObjs, err := loadRTEManifests()
	if err != nil {
		return err
	}
	objs = append(objs, rteObjs...)

	schedObjs, err := loadSchedPluginManifests()
	if err != nil {
		return err
	}
	objs = append(objs, schedObjs...)

	for _, obj := range objs {
		fmt.Printf("---\n")
		if err := manifests.SerializeObject(obj, os.Stdout); err != nil {
			return err
		}
	}

	return nil
}

func loadRTEManifests() ([]runtime.Object, error) {
	var objs []runtime.Object

	ns, err := manifests.LoadNamespace(manifests.ComponentResourceTopologyExporter)
	if err != nil {
		return nil, err
	}
	objs = append(objs, ns)

	sa, err := manifests.LoadServiceAccount(manifests.ComponentResourceTopologyExporter)
	if err != nil {
		return nil, err
	}
	objs = append(objs, sa)

	cr, err := manifests.LoadClusterRole(manifests.ComponentResourceTopologyExporter)
	if err != nil {
		return nil, err
	}
	objs = append(objs, cr)

	crb, err := manifests.LoadResourceTopologyExporterClusterRoleBinding()
	if err != nil {
		return nil, err
	}
	objs = append(objs, crb)

	ds, err := manifests.LoadResourceTopologyExporterDaemonSet()
	if err != nil {
		return nil, err
	}
	objs = append(objs, manifests.UpdateResourceTopologyExporterDaemonSet(ds))

	return objs, nil
}

type loadSchedCRBFunc func() (*rbacv1.ClusterRoleBinding, error)

func loadSchedPluginManifests() ([]runtime.Object, error) {
	var objs []runtime.Object

	ns, err := manifests.LoadNamespace(manifests.ComponentSchedulerPlugin)
	if err != nil {
		return nil, err
	}
	objs = append(objs, ns)

	sa, err := manifests.LoadServiceAccount(manifests.ComponentSchedulerPlugin)
	if err != nil {
		return nil, err
	}
	objs = append(objs, sa)

	cr, err := manifests.LoadClusterRole(manifests.ComponentSchedulerPlugin)
	if err != nil {
		return nil, err
	}
	objs = append(objs, cr)

	for _, loader := range []loadSchedCRBFunc{
		manifests.LoadSchedulerPluginClusterRoleBindingKubeScheduler,
		manifests.LoadSchedulerPluginClusterRoleBindingNodeResourceTopology,
		manifests.LoadSchedulerPluginClusterRoleBindingVolumeScheduler,
	} {
		crb, err := loader()
		if err != nil {
			return nil, err
		}
		objs = append(objs, crb)

	}

	rb, err := manifests.LoadSchedulerPluginRoleBindingKubeScheduler()
	if err != nil {
		return nil, err
	}
	objs = append(objs, rb)

	cm, err := manifests.LoadSchedulerPluginConfigMap()
	if err != nil {
		return nil, err
	}
	objs = append(objs, cm)

	dp, err := manifests.LoadSchedulerPluginDeployment()
	if err != nil {
		return nil, err
	}
	objs = append(objs, manifests.UpdateSchedulerPluginDeployment(dp))

	return objs, nil
}
