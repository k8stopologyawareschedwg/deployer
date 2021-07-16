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

package rte

import (
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/fromanirh/deployer/pkg/deployer"
	rtemanifests "github.com/fromanirh/deployer/pkg/manifests/rte"
)

type Options struct{}

func Deploy(log deployer.Logger, opts Options) error {
	var err error
	log.Printf("deploying topology-aware-scheduling topology updater...")

	mf, err := rtemanifests.GetManifests()
	if err != nil {
		return err
	}
	mf = mf.UpdateNamespace().UpdatePullspecs()
	log.Debugf("RTE manifests loaded")

	hp, err := deployer.NewHelper("RTE", log)
	if err != nil {
		return err
	}

	for _, obj := range mf.ToObjects() {
		if err := hp.CreateObject(obj); err != nil {
			return err
		}
	}

	dsKey := types.NamespacedName{
		Name:      mf.DaemonSet.Name,
		Namespace: mf.Namespace.Name,
	}

	if err = hp.WaitForObjectToBeCreated(dsKey, &appsv1.DaemonSet{}); err != nil {
		return err
	}

	log.Printf("...deployed topology-aware-scheduling topology updater!")
	return nil
}

func Remove(log deployer.Logger, opts Options) error {
	var err error
	log.Printf("removing topology-aware-scheduling topology updater...")

	hp, err := deployer.NewHelper("RTE", log)
	if err != nil {
		return err
	}

	mf, err := rtemanifests.GetManifests()
	if err != nil {
		return err
	}
	mf = mf.UpdateNamespace()
	log.Debugf("RTE manifests loaded")

	// since we created everything in the namespace, we can just do
	if err := hp.DeleteObject(mf.Namespace); err != nil {
		return err
	}

	// TODO: (optional) wait for the namespace to be gone

	// but now let's take care of cluster-scoped resources
	if err := hp.DeleteObject(mf.ClusterRoleBinding); err != nil {
		return err
	}
	if err := hp.DeleteObject(mf.ClusterRole); err != nil {
		return err
	}

	log.Printf("...removed topology-aware-scheduling topology updater!")
	return nil
}
