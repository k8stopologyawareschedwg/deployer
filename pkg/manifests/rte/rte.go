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
	"github.com/fromanirh/deployer/pkg/manifests"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Manifests struct {
	Namespace          *corev1.Namespace
	ServiceAccount     *corev1.ServiceAccount
	ClusterRole        *rbacv1.ClusterRole
	ClusterRoleBinding *rbacv1.ClusterRoleBinding
	DaemonSet          *appsv1.DaemonSet
}

func (mf Manifests) UpdateNamespace() Manifests {
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
