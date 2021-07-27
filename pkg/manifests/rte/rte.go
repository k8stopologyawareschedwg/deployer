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
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fromanirh/deployer/pkg/deployer"
	"github.com/fromanirh/deployer/pkg/deployer/platform"
	"github.com/fromanirh/deployer/pkg/deployer/wait"
	"github.com/fromanirh/deployer/pkg/manifests"
)

type Manifests struct {
	Namespace          *corev1.Namespace
	ServiceAccount     *corev1.ServiceAccount
	ClusterRole        *rbacv1.ClusterRole
	ClusterRoleBinding *rbacv1.ClusterRoleBinding
	DaemonSet          *appsv1.DaemonSet
	plat               platform.Platform
}

func (mf Manifests) Clone() Manifests {
	return Manifests{
		Namespace:          mf.Namespace.DeepCopy(),
		ServiceAccount:     mf.ServiceAccount.DeepCopy(),
		ClusterRole:        mf.ClusterRole.DeepCopy(),
		ClusterRoleBinding: mf.ClusterRoleBinding.DeepCopy(),
		DaemonSet:          mf.DaemonSet.DeepCopy(),
	}
}

func (mf Manifests) UpdateNamespace() Manifests {
	ret := mf.Clone()
	ret.ServiceAccount.Namespace = ret.Namespace.Name
	ret.DaemonSet.Namespace = ret.Namespace.Name
	for idx := 0; idx < len(ret.ClusterRoleBinding.Subjects); idx++ {
		ret.ClusterRoleBinding.Subjects[idx].Namespace = ret.Namespace.Name
	}
	return ret
}

func (mf Manifests) UpdatePullspecs() Manifests {
	ret := mf.Clone()
	ret.DaemonSet = manifests.UpdateResourceTopologyExporterDaemonSet(ret.DaemonSet)
	return ret
}

func (mf Manifests) ToObjects() []client.Object {
	return []client.Object{
		mf.Namespace,
		mf.ServiceAccount,
		mf.ClusterRole,
		mf.ClusterRoleBinding,
		mf.DaemonSet,
	}
}

func (mf Manifests) ToCreatableObjects(hp *deployer.Helper, log deployer.Logger) []deployer.WaitableObject {
	return []deployer.WaitableObject{
		deployer.WaitableObject{Obj: mf.Namespace},
		deployer.WaitableObject{Obj: mf.ServiceAccount},
		deployer.WaitableObject{Obj: mf.ClusterRole},
		deployer.WaitableObject{Obj: mf.ClusterRoleBinding},
		deployer.WaitableObject{
			Obj:  mf.DaemonSet,
			Wait: func() error { return wait.PodsToBeRunningByRegex(hp, log, mf.DaemonSet.Namespace, mf.DaemonSet.Name) },
		},
	}

}

func (mf Manifests) ToDeletableObjects(hp *deployer.Helper, log deployer.Logger) []deployer.WaitableObject {
	return []deployer.WaitableObject{
		deployer.WaitableObject{
			Obj:  mf.Namespace,
			Wait: func() error { return wait.NamespaceToBeGone(hp, log, mf.Namespace.Name) },
		},
		// no need to remove objects created inside the namespace we just removed
		deployer.WaitableObject{Obj: mf.ClusterRoleBinding},
		deployer.WaitableObject{Obj: mf.ClusterRole},
	}
}

func GetManifests(plat platform.Platform) (Manifests, error) {
	var err error
	mf := Manifests{
		plat: plat,
	}
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
