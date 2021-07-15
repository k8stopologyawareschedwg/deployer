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

package sched

import (
	"github.com/fromanirh/deployer/pkg/manifests"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Manifests struct {
	Namespace               *corev1.Namespace
	ServiceAccount          *corev1.ServiceAccount
	ClusterRole             *rbacv1.ClusterRole
	CRBKubernetesScheduler  *rbacv1.ClusterRoleBinding
	CRBNodeResourceTopology *rbacv1.ClusterRoleBinding
	CRBVolumeScheduler      *rbacv1.ClusterRoleBinding
	RoleBinding             *rbacv1.RoleBinding
	ConfigMap               *corev1.ConfigMap
	Deployment              *appsv1.Deployment
}

func (mf Manifests) Clone() Manifests {
	return Manifests{
		Namespace:               mf.Namespace.DeepCopy(),
		ServiceAccount:          mf.ServiceAccount.DeepCopy(),
		ClusterRole:             mf.ClusterRole.DeepCopy(),
		CRBKubernetesScheduler:  mf.CRBKubernetesScheduler.DeepCopy(),
		CRBNodeResourceTopology: mf.CRBNodeResourceTopology.DeepCopy(),
		CRBVolumeScheduler:      mf.CRBVolumeScheduler.DeepCopy(),
		RoleBinding:             mf.RoleBinding.DeepCopy(),
		ConfigMap:               mf.ConfigMap.DeepCopy(),
		Deployment:              mf.Deployment.DeepCopy(),
	}
}

func (mf Manifests) UpdateNamespace() Manifests {
	ret := mf.Clone()
	ret.ServiceAccount.Namespace = mf.Namespace.Name
	ret.ConfigMap.Namespace = mf.Namespace.Name
	ret.Deployment.Namespace = mf.Namespace.Name
	return ret
}

func (mf Manifests) UpdatePullspecs() Manifests {
	ret := mf.Clone()
	ret.Deployment = manifests.UpdateSchedulerPluginDeployment(ret.Deployment)
	return ret
}

func (mf Manifests) ToObjects() []runtime.Object {
	return []runtime.Object{
		mf.Namespace,
		mf.ServiceAccount,
		mf.ClusterRole,
		mf.CRBKubernetesScheduler,
		mf.CRBNodeResourceTopology,
		mf.CRBVolumeScheduler,
		mf.RoleBinding,
		mf.ConfigMap,
		mf.Deployment,
	}
}

func GetManifests() (Manifests, error) {
	var err error
	mf := Manifests{}
	mf.Namespace, err = manifests.Namespace(manifests.ComponentSchedulerPlugin)
	if err != nil {
		return mf, err
	}
	mf.ServiceAccount, err = manifests.ServiceAccount(manifests.ComponentSchedulerPlugin)
	if err != nil {
		return mf, err
	}
	mf.ClusterRole, err = manifests.ClusterRole(manifests.ComponentSchedulerPlugin)
	if err != nil {
		return mf, err
	}
	mf.CRBKubernetesScheduler, err = manifests.SchedulerPluginClusterRoleBindingKubeScheduler()
	if err != nil {
		return mf, err
	}
	mf.CRBNodeResourceTopology, err = manifests.SchedulerPluginClusterRoleBindingNodeResourceTopology()
	if err != nil {
		return mf, err
	}
	mf.CRBVolumeScheduler, err = manifests.SchedulerPluginClusterRoleBindingVolumeScheduler()
	if err != nil {
		return mf, err
	}
	mf.RoleBinding, err = manifests.SchedulerPluginRoleBindingKubeScheduler()
	if err != nil {
		return mf, err
	}
	mf.ConfigMap, err = manifests.SchedulerPluginConfigMap()
	if err != nil {
		return mf, err
	}
	mf.Deployment, err = manifests.SchedulerPluginDeployment()
	if err != nil {
		return mf, err
	}
	return mf, nil
}
