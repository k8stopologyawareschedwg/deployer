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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fromanirh/deployer/pkg/deployer"
	"github.com/fromanirh/deployer/pkg/deployer/platform"
	"github.com/fromanirh/deployer/pkg/deployer/wait"
	"github.com/fromanirh/deployer/pkg/manifests"
)

type Manifests struct {
	ServiceAccount          *corev1.ServiceAccount
	ClusterRole             *rbacv1.ClusterRole
	CRBKubernetesScheduler  *rbacv1.ClusterRoleBinding
	CRBNodeResourceTopology *rbacv1.ClusterRoleBinding
	CRBVolumeScheduler      *rbacv1.ClusterRoleBinding
	RoleBinding             *rbacv1.RoleBinding
	ConfigMap               *corev1.ConfigMap
	Deployment              *appsv1.Deployment
	plat                    platform.Platform
}

func (mf Manifests) Clone() Manifests {
	return Manifests{
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

func (mf Manifests) Update() Manifests {
	ret := mf.Clone()
	ret.Deployment = manifests.UpdateSchedulerPluginDeployment(ret.Deployment)
	return ret
}

func (mf Manifests) ToObjects() []client.Object {
	return []client.Object{
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

func (mf Manifests) ToCreatableObjects(hp *deployer.Helper, log deployer.Logger) []deployer.WaitableObject {
	return []deployer.WaitableObject{
		deployer.WaitableObject{Obj: mf.ServiceAccount},
		deployer.WaitableObject{Obj: mf.ClusterRole},
		deployer.WaitableObject{Obj: mf.CRBKubernetesScheduler},
		deployer.WaitableObject{Obj: mf.CRBNodeResourceTopology},
		deployer.WaitableObject{Obj: mf.CRBVolumeScheduler},
		deployer.WaitableObject{Obj: mf.RoleBinding},
		deployer.WaitableObject{Obj: mf.ConfigMap},
		deployer.WaitableObject{
			Obj:  mf.Deployment,
			Wait: func() error { return wait.PodsToBeRunningByRegex(hp, log, mf.Deployment.Namespace, mf.Deployment.Name) },
		},
	}
}

func (mf Manifests) ToDeletableObjects(hp *deployer.Helper, log deployer.Logger) []deployer.WaitableObject {
	return []deployer.WaitableObject{
		deployer.WaitableObject{
			Obj:  mf.Deployment,
			Wait: func() error { return wait.PodsToBeGoneByRegex(hp, log, mf.Deployment.Namespace, mf.Deployment.Name) },
		},
		deployer.WaitableObject{Obj: mf.ConfigMap},
		deployer.WaitableObject{Obj: mf.RoleBinding},
		deployer.WaitableObject{Obj: mf.CRBVolumeScheduler},
		deployer.WaitableObject{Obj: mf.CRBNodeResourceTopology},
		deployer.WaitableObject{Obj: mf.CRBKubernetesScheduler},
		deployer.WaitableObject{Obj: mf.ClusterRole},
		deployer.WaitableObject{Obj: mf.ServiceAccount},
	}
}

func GetManifests(plat platform.Platform) (Manifests, error) {
	var err error
	mf := Manifests{
		plat: plat,
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
