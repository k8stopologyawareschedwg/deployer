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
	"github.com/go-logr/logr"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/wait"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
)

const (
	NamespaceOpenShift = "openshift-topology-aware-scheduler"
)

type Manifests struct {
	// common
	Crd       *apiextensionv1.CustomResourceDefinition
	Namespace *corev1.Namespace
	// controller
	SAController  *corev1.ServiceAccount
	CRController  *rbacv1.ClusterRole
	CRBController *rbacv1.ClusterRoleBinding
	RBController  *rbacv1.RoleBinding
	DPController  *appsv1.Deployment
	// scheduler proper
	SAScheduler  *corev1.ServiceAccount
	CRScheduler  *rbacv1.ClusterRole
	CRBScheduler *rbacv1.ClusterRoleBinding
	RBScheduler  *rbacv1.RoleBinding
	DPScheduler  *appsv1.Deployment
	ConfigMap    *corev1.ConfigMap
	// internal fields
	plat platform.Platform
}

func (mf Manifests) Clone() Manifests {
	return Manifests{
		plat: mf.plat,
		// objects
		Crd:           mf.Crd.DeepCopy(),
		Namespace:     mf.Namespace.DeepCopy(),
		SAController:  mf.SAController.DeepCopy(),
		CRController:  mf.CRController.DeepCopy(),
		CRBController: mf.CRBController.DeepCopy(),
		DPController:  mf.DPController.DeepCopy(),
		RBController:  mf.RBController.DeepCopy(),
		SAScheduler:   mf.SAScheduler.DeepCopy(),
		CRScheduler:   mf.CRScheduler.DeepCopy(),
		CRBScheduler:  mf.CRBScheduler.DeepCopy(),
		DPScheduler:   mf.DPScheduler.DeepCopy(),
		ConfigMap:     mf.ConfigMap.DeepCopy(),
		RBScheduler:   mf.RBScheduler.DeepCopy(),
	}
}

type RenderOptions struct {
	Replicas         int32
	PullIfNotPresent bool
}

func (mf Manifests) Render(logger logr.Logger, options RenderOptions) Manifests {
	ret := mf.Clone()
	replicas := options.Replicas
	if replicas <= 0 {
		replicas = int32(1)
	}
	ret.DPScheduler.Spec.Replicas = newInt32(replicas)
	ret.DPController.Spec.Replicas = newInt32(replicas)

	manifests.UpdateSchedulerPluginSchedulerDeployment(ret.DPScheduler, options.PullIfNotPresent)
	manifests.UpdateSchedulerPluginControllerDeployment(ret.DPController, options.PullIfNotPresent)
	if mf.plat == platform.OpenShift {
		ret.Namespace.Name = NamespaceOpenShift
	}

	ret.SAController.Namespace = ret.Namespace.Name
	manifests.UpdateClusterRoleBinding(ret.CRBController, ret.SAController.Name, ret.Namespace.Name)
	manifests.UpdateRoleBinding(ret.RBController, ret.SAController.Name, ret.Namespace.Name)
	ret.DPController.Namespace = ret.Namespace.Name

	ret.SAScheduler.Namespace = ret.Namespace.Name
	manifests.UpdateClusterRoleBinding(ret.CRBScheduler, ret.SAScheduler.Name, ret.Namespace.Name)
	manifests.UpdateRoleBinding(ret.RBScheduler, ret.SAScheduler.Name, ret.Namespace.Name)
	ret.DPScheduler.Namespace = ret.Namespace.Name
	ret.ConfigMap.Namespace = ret.Namespace.Name

	return ret
}

func (mf Manifests) ToObjects() []client.Object {
	return []client.Object{
		mf.Crd,
		mf.Namespace,
		mf.SAScheduler,
		mf.CRScheduler,
		mf.CRBScheduler,
		mf.ConfigMap,
		mf.RBScheduler,
		mf.DPScheduler,
		mf.SAController,
		mf.CRController,
		mf.CRBController,
		mf.DPController,
		mf.RBController,
	}
}

func (mf Manifests) ToCreatableObjects(hp *deployer.Helper, log logr.Logger) []deployer.WaitableObject {
	return []deployer.WaitableObject{
		{Obj: mf.Crd},
		{Obj: mf.Namespace},
		{Obj: mf.SAScheduler},
		{Obj: mf.CRScheduler},
		{Obj: mf.CRBScheduler},
		{Obj: mf.RBScheduler},
		{Obj: mf.ConfigMap},
		{
			Obj: mf.DPScheduler,
			Wait: func() error {
				return wait.PodsToBeRunningByRegex(hp, log, mf.DPScheduler.Namespace, mf.DPScheduler.Name)
			},
		},
		{Obj: mf.SAController},
		{Obj: mf.CRController},
		{Obj: mf.CRBController},
		{Obj: mf.RBController},
		{
			Obj: mf.DPController,
			Wait: func() error {
				return wait.PodsToBeRunningByRegex(hp, log, mf.DPController.Namespace, mf.DPController.Name)
			},
		},
	}
}

func (mf Manifests) ToDeletableObjects(hp *deployer.Helper, log logr.Logger) []deployer.WaitableObject {
	return []deployer.WaitableObject{
		{
			Obj:  mf.Namespace,
			Wait: func() error { return wait.NamespaceToBeGone(hp, log, mf.Namespace.Name) },
		},
		// no need to remove objects created inside the namespace we just removed
		{Obj: mf.CRBScheduler},
		{Obj: mf.CRScheduler},
		{Obj: mf.RBScheduler},
		{Obj: mf.CRBController},
		{Obj: mf.CRController},
		{Obj: mf.RBController},
		{Obj: mf.Crd},
	}
}

func New(plat platform.Platform) Manifests {
	return Manifests{
		plat: plat,
	}
}

func GetManifests(plat platform.Platform, namespace string) (Manifests, error) {
	var err error
	mf := New(plat)
	mf.Crd, err = manifests.SchedulerCRD()
	if err != nil {
		return mf, err
	}
	mf.Namespace, err = manifests.Namespace(manifests.ComponentSchedulerPlugin)
	if err != nil {
		return mf, err
	}

	mf.ConfigMap, err = manifests.ConfigMap(manifests.ComponentSchedulerPlugin, "")
	if err != nil {
		return mf, err
	}
	mf.SAScheduler, err = manifests.ServiceAccount(manifests.ComponentSchedulerPlugin, manifests.SubComponentSchedulerPluginScheduler, namespace)
	if err != nil {
		return mf, err
	}
	mf.CRScheduler, err = manifests.ClusterRole(manifests.ComponentSchedulerPlugin, manifests.SubComponentSchedulerPluginScheduler)
	if err != nil {
		return mf, err
	}
	mf.CRBScheduler, err = manifests.ClusterRoleBinding(manifests.ComponentSchedulerPlugin, manifests.SubComponentSchedulerPluginScheduler)
	if err != nil {
		return mf, err
	}
	mf.RBScheduler, err = manifests.RoleBinding(manifests.ComponentSchedulerPlugin, manifests.SubComponentSchedulerPluginScheduler, namespace)
	if err != nil {
		return mf, err
	}
	mf.DPScheduler, err = manifests.Deployment(manifests.ComponentSchedulerPlugin, manifests.SubComponentSchedulerPluginScheduler, "")
	if err != nil {
		return mf, err
	}

	mf.SAController, err = manifests.ServiceAccount(manifests.ComponentSchedulerPlugin, manifests.SubComponentSchedulerPluginController, namespace)
	if err != nil {
		return mf, err
	}
	mf.CRController, err = manifests.ClusterRole(manifests.ComponentSchedulerPlugin, manifests.SubComponentSchedulerPluginController)
	if err != nil {
		return mf, err
	}
	mf.CRBController, err = manifests.ClusterRoleBinding(manifests.ComponentSchedulerPlugin, manifests.SubComponentSchedulerPluginController)
	if err != nil {
		return mf, err
	}
	mf.RBController, err = manifests.RoleBinding(manifests.ComponentSchedulerPlugin, manifests.SubComponentSchedulerPluginController, namespace)
	if err != nil {
		return mf, err
	}
	mf.DPController, err = manifests.Deployment(manifests.ComponentSchedulerPlugin, manifests.SubComponentSchedulerPluginController, "")
	if err != nil {
		return mf, err
	}

	return mf, nil
}

func newInt32(value int32) *int32 {
	return &value
}
