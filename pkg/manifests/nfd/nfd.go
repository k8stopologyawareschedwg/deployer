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
 * Copyright 2022 Red Hat, Inc.
 */

package nfd

import (
	"github.com/go-logr/logr"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/wait"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
)

type Manifests struct {
	Namespace *corev1.Namespace
	// master objects
	SAMaster  *corev1.ServiceAccount
	CRMaster  *rbacv1.ClusterRole
	CRBMaster *rbacv1.ClusterRoleBinding
	DPMaster  *appsv1.Deployment
	SVMaster  *corev1.Service

	// topology-updater objects
	SATopologyUpdater  *corev1.ServiceAccount
	CRTopologyUpdater  *rbacv1.ClusterRole
	CRBTopologyUpdater *rbacv1.ClusterRoleBinding
	DSTopologyUpdater  *appsv1.DaemonSet

	plat platform.Platform
}

func (mf Manifests) Clone() Manifests {
	ret := Manifests{
		plat: mf.plat,

		Namespace: mf.Namespace.DeepCopy(),
		// master objects
		CRMaster:  mf.CRMaster.DeepCopy(),
		CRBMaster: mf.CRBMaster.DeepCopy(),
		DPMaster:  mf.DPMaster.DeepCopy(),
		SAMaster:  mf.SAMaster.DeepCopy(),
		SVMaster:  mf.SVMaster.DeepCopy(),

		// topology-updater objects
		CRTopologyUpdater:  mf.CRTopologyUpdater.DeepCopy(),
		CRBTopologyUpdater: mf.CRBTopologyUpdater.DeepCopy(),
		DSTopologyUpdater:  mf.DSTopologyUpdater.DeepCopy(),
		SATopologyUpdater:  mf.SATopologyUpdater.DeepCopy(),
	}

	return ret
}

type RenderOptions struct {
	PullIfNotPresent bool
	// DaemonSet option
	NodeSelector *metav1.LabelSelector
	// Deployment option
	Replicas int32

	// General options
	Namespace string
}

func (mf Manifests) Render(options RenderOptions) Manifests {
	ret := mf.Clone()

	replicas := options.Replicas
	if replicas <= 0 {
		replicas = int32(1)
	}
	ret.DPMaster.Spec.Replicas = &replicas

	if options.Namespace != "" {
		ret.Namespace.Name = options.Namespace
	}

	manifests.UpdateClusterRoleBinding(ret.CRBMaster, mf.SAMaster.Name, ret.Namespace.Name)
	manifests.UpdateClusterRoleBinding(ret.CRBTopologyUpdater, mf.SATopologyUpdater.Name, ret.Namespace.Name)

	ret.DPMaster.Spec.Template.Spec.ServiceAccountName = mf.SAMaster.Name
	ret.DSTopologyUpdater.Spec.Template.Spec.ServiceAccountName = mf.SATopologyUpdater.Name

	manifests.UpdateNFDMasterDeployment(ret.DPMaster, options.PullIfNotPresent)
	manifests.UpdateNFDTopologyUpdaterDaemonSet(ret.DSTopologyUpdater, options.PullIfNotPresent, options.NodeSelector)

	return ret
}

func (mf Manifests) ToObjects() []client.Object {
	return []client.Object{
		mf.Namespace,
		// master objects
		mf.CRMaster,
		mf.CRBMaster,
		mf.SAMaster,
		mf.DPMaster,
		mf.SVMaster,
		// topology-updater objects
		mf.SATopologyUpdater,
		mf.CRTopologyUpdater,
		mf.CRBTopologyUpdater,
		mf.DSTopologyUpdater,
	}
}

func (mf Manifests) ToCreatableObjects(cli client.Client, log logr.Logger) []deployer.WaitableObject {
	return []deployer.WaitableObject{
		{Obj: mf.CRMaster},
		{Obj: mf.CRBMaster},
		{Obj: mf.SAMaster},
		{Obj: mf.SVMaster},
		{
			Obj: mf.DPMaster,
			Wait: func() error {
				return wait.PodsToBeRunningByRegex(cli, log, mf.DPMaster.Namespace, mf.DPMaster.Name)
			},
		},
		{Obj: mf.SATopologyUpdater},
		{Obj: mf.CRTopologyUpdater},
		{Obj: mf.CRBTopologyUpdater},
		{
			Obj: mf.DSTopologyUpdater,
			Wait: func() error {
				return wait.PodsToBeRunningByRegex(cli, log, mf.DSTopologyUpdater.Namespace, mf.DSTopologyUpdater.Name)
			},
		},
	}
}

func (mf Manifests) ToDeletableObjects(cli client.Client, log logr.Logger) []deployer.WaitableObject {
	return []deployer.WaitableObject{
		{Obj: mf.CRBTopologyUpdater},
		{Obj: mf.CRTopologyUpdater},
		{Obj: mf.CRBMaster},
		{Obj: mf.CRMaster},
	}
}

func New(plat platform.Platform) Manifests {
	mf := Manifests{
		plat: plat,
	}

	return mf
}

func GetManifests(plat platform.Platform, namespace string) (Manifests, error) {
	var err error
	mf := New(plat)

	mf.Namespace, err = manifests.Namespace(manifests.ComponentNodeFeatureDiscovery)
	if err != nil {
		return mf, err
	}

	mf.SAMaster, err = manifests.ServiceAccount(manifests.ComponentNodeFeatureDiscovery, manifests.SubComponentNodeFeatureDiscoveryMaster, namespace)
	if err != nil {
		return mf, err
	}
	mf.CRMaster, err = manifests.ClusterRole(manifests.ComponentNodeFeatureDiscovery, manifests.SubComponentNodeFeatureDiscoveryMaster)
	if err != nil {
		return mf, err
	}
	mf.CRBMaster, err = manifests.ClusterRoleBinding(manifests.ComponentNodeFeatureDiscovery, manifests.SubComponentNodeFeatureDiscoveryMaster)
	if err != nil {
		return mf, err
	}
	mf.DPMaster, err = manifests.Deployment(manifests.ComponentNodeFeatureDiscovery, manifests.SubComponentNodeFeatureDiscoveryMaster, namespace)
	if err != nil {
		return mf, err
	}
	mf.SATopologyUpdater, err = manifests.ServiceAccount(manifests.ComponentNodeFeatureDiscovery, manifests.SubComponentNodeFeatureDiscoveryTopologyUpdater, namespace)
	if err != nil {
		return mf, err
	}
	mf.CRTopologyUpdater, err = manifests.ClusterRole(manifests.ComponentNodeFeatureDiscovery, manifests.SubComponentNodeFeatureDiscoveryTopologyUpdater)
	if err != nil {
		return mf, err
	}
	mf.CRBTopologyUpdater, err = manifests.ClusterRoleBinding(manifests.ComponentNodeFeatureDiscovery, manifests.SubComponentNodeFeatureDiscoveryTopologyUpdater)
	if err != nil {
		return mf, err
	}
	mf.DSTopologyUpdater, err = manifests.DaemonSet(manifests.ComponentNodeFeatureDiscovery, manifests.SubComponentNodeFeatureDiscoveryTopologyUpdater, plat, namespace)
	if err != nil {
		return mf, err
	}
	mf.SVMaster, err = manifests.Service(manifests.ComponentNodeFeatureDiscovery, manifests.SubComponentNodeFeatureDiscoveryMaster, namespace)
	if err != nil {
		return mf, err
	}

	return mf, nil
}
