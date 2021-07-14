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
	"log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/fromanirh/deployer/pkg/deployer"
	"github.com/fromanirh/deployer/pkg/manifests"
)

type Options struct{}

type Manifests struct {
	Namespace          *corev1.Namespace
	ServiceAccount     *corev1.ServiceAccount
	ClusterRole        *rbacv1.ClusterRole
	ClusterRoleBinding *rbacv1.ClusterRoleBinding
	DaemonSet          *appsv1.DaemonSet
}

func (mf Manifests) EnforceNamespace() Manifests {
	ret := Manifests{
		Namespace:          mf.Namespace.DeepCopy(),
		ServiceAccount:     mf.ServiceAccount.DeepCopy(),
		ClusterRole:        mf.ClusterRole.DeepCopy(),
		ClusterRoleBinding: mf.ClusterRoleBinding.DeepCopy(),
		DaemonSet:          mf.DaemonSet.DeepCopy(),
	}
	ret.ServiceAccount.Namespace = ret.Namespace.Name
	ret.DaemonSet.Namespace = ret.Namespace.Name
	return ret
}

func (mf Manifests) ToObjects() []runtime.Object {
	return []runtime.Object{
		mf.Namespace,
		mf.ServiceAccount,
		mf.ClusterRole,
		mf.ClusterRoleBinding,
		mf.DaemonSet,
	}
}

func GetManifests() (Manifests, error) {
	var err error
	mf := Manifests{}
	mf.Namespace, err = manifests.Namespace(manifests.ComponentResourceTopologyExporter)
	if err != nil {
		return mf, err
	}
	mf.ServiceAccount, err = manifests.ServiceAccount(manifests.ComponentResourceTopologyExporter)
	if err != nil {
		return mf, err
	}
	mf.ClusterRole, err = manifests.ClusterRole(manifests.ComponentResourceTopologyExporter)
	if err != nil {
		return mf, err
	}
	mf.ClusterRoleBinding, err = manifests.ResourceTopologyExporterClusterRoleBinding()
	if err != nil {
		return mf, err
	}
	mf.DaemonSet, err = manifests.ResourceTopologyExporterDaemonSet()
	if err != nil {
		return mf, err
	}
	return mf, nil
}

func Deploy(logger *log.Logger, opts Options) error {
	var err error

	mf, err := GetManifests()
	if err != nil {
		return err
	}
	mf = mf.EnforceNamespace()
	logger.Printf("manifests loaded")

	hp, err := deployer.NewHelper("RTE")
	if err != nil {
		return err
	}

	if err := hp.CreateObject(mf.Namespace); err != nil {
		return err
	}

	if err := hp.CreateObject(mf.ServiceAccount); err != nil {
		return err
	}
	if err := hp.CreateObject(mf.ClusterRole); err != nil {
		return err
	}
	if err := hp.CreateObject(mf.ClusterRoleBinding); err != nil {
		return err
	}
	ds := manifests.UpdateResourceTopologyExporterDaemonSet(mf.DaemonSet)
	if err := hp.CreateObject(ds); err != nil {
		return err
	}

	// TODO: (optional) wait for the DS to go running

	return nil
}

func Remove(logger *log.Logger, opts Options) error {
	var err error

	hp, err := deployer.NewHelper("RTE")
	if err != nil {
		return err
	}

	mf, err := GetManifests()
	if err != nil {
		return err
	}
	mf = mf.EnforceNamespace()
	logger.Printf("manifests loaded")

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

	return nil
}
