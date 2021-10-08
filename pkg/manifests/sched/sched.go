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
	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/wait"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
	"github.com/k8stopologyawareschedwg/deployer/pkg/tlog"
)

const (
	namespaceOCP = "openshift-topology-aware-scheduler"
)

type Manifests struct {
	// common
	Crd       *apiextensionv1.CustomResourceDefinition
	Namespace *corev1.Namespace
	// controller
	SAController  *corev1.ServiceAccount
	CRController  *rbacv1.ClusterRole
	CRBController *rbacv1.ClusterRoleBinding
	DPController  *appsv1.Deployment
	// scheduler proper
	SAScheduler  *corev1.ServiceAccount
	CRScheduler  *rbacv1.ClusterRole
	CRBScheduler *rbacv1.ClusterRoleBinding
	DPScheduler  *appsv1.Deployment
	ConfigMap    *corev1.ConfigMap
	// internal fields
	plat      platform.Platform
	namespace string
}

func (mf Manifests) Clone() Manifests {
	ret := Manifests{
		plat:      mf.plat,
		namespace: mf.namespace,
		// objects
		Crd:           mf.Crd.DeepCopy(),
		SAController:  mf.SAController.DeepCopy(),
		CRController:  mf.CRController.DeepCopy(),
		CRBController: mf.CRBController.DeepCopy(),
		DPController:  mf.DPController.DeepCopy(),
		SAScheduler:   mf.SAScheduler.DeepCopy(),
		CRScheduler:   mf.CRScheduler.DeepCopy(),
		CRBScheduler:  mf.CRBScheduler.DeepCopy(),
		DPScheduler:   mf.DPScheduler.DeepCopy(),
		ConfigMap:     mf.ConfigMap.DeepCopy(),
	}
	if mf.Namespace != nil {
		ret.Namespace = mf.Namespace.DeepCopy()
	}
	return ret
}

type UpdateOptions struct {
	Replicas               int32
	NodeResourcesNamespace string
	PullIfNotPresent       bool
}

func (mf Manifests) Update(logger tlog.Logger, options UpdateOptions) Manifests {
	ret := mf.Clone()

	if mf.plat == platform.OpenShift {
		mf.Namespace.Name = namespaceOCP
	}
	mf.namespace = mf.Namespace.Name

	replicas := options.Replicas
	if replicas <= 0 {
		replicas = int32(1)
	}
	ret.DPScheduler.Spec.Replicas = newInt32(replicas)
	ret.DPController.Spec.Replicas = newInt32(replicas)

	manifests.UpdateSchedulerPluginSchedulerDeployment(ret.DPScheduler, options.PullIfNotPresent)
	manifests.UpdateSchedulerPluginControllerDeployment(ret.DPController, options.PullIfNotPresent)

	ret.SAController.Namespace = mf.namespace
	manifests.UpdateClusterRoleBinding(ret.CRBController, "", mf.namespace)
	ret.DPController.Namespace = mf.namespace

	ret.SAScheduler.Namespace = mf.namespace
	manifests.UpdateClusterRoleBinding(ret.CRBScheduler, "", mf.namespace)
	ret.DPScheduler.Namespace = mf.namespace
	ret.ConfigMap.Namespace = mf.namespace

	if options.NodeResourcesNamespace != "" {
		ret.ConfigMap = manifests.UpdateSchedulerConfigNamespaces(logger, ret.ConfigMap, options.NodeResourcesNamespace)
	}
	return ret
}

func (mf Manifests) ToObjects() []client.Object {
	objs := []client.Object{
		mf.Crd,
	}
	if mf.Namespace != nil {
		objs = append(objs, mf.Namespace)
	}
	objs = append(objs,
		mf.Namespace,
		mf.SAScheduler,
		mf.CRScheduler,
		mf.CRBScheduler,
		mf.ConfigMap,
		mf.DPScheduler,
		mf.SAController,
		mf.CRController,
		mf.CRBController,
		mf.DPController,
	)
	return objs
}

func (mf Manifests) ToCreatableObjects(hp *deployer.Helper, log tlog.Logger) []deployer.WaitableObject {
	objs := []deployer.WaitableObject{
		{Obj: mf.Crd},
	}
	if mf.Namespace != nil {
		objs = append(objs, deployer.WaitableObject{Obj: mf.Namespace})
	}
	objs = append(objs,
		deployer.WaitableObject{Obj: mf.SAScheduler},
		deployer.WaitableObject{Obj: mf.CRScheduler},
		deployer.WaitableObject{Obj: mf.CRBScheduler},
		deployer.WaitableObject{Obj: mf.ConfigMap},
		deployer.WaitableObject{
			Obj: mf.DPScheduler,
			Wait: func() error {
				return wait.PodsToBeRunningByRegex(hp, log, mf.DPScheduler.Namespace, mf.DPScheduler.Name)
			},
		},
		deployer.WaitableObject{Obj: mf.SAController},
		deployer.WaitableObject{Obj: mf.CRController},
		deployer.WaitableObject{Obj: mf.CRBController},
		deployer.WaitableObject{
			Obj: mf.DPController,
			Wait: func() error {
				return wait.PodsToBeRunningByRegex(hp, log, mf.DPController.Namespace, mf.DPController.Name)
			},
		},
	)
	return objs
}

func (mf Manifests) ToDeletableObjects(hp *deployer.Helper, log tlog.Logger) []deployer.WaitableObject {
	if mf.Namespace != nil {
		return []deployer.WaitableObject{
			{
				Obj:  mf.Namespace,
				Wait: func() error { return wait.NamespaceToBeGone(hp, log, mf.Namespace.Name) },
			},
			// no need to remove objects created inside the namespace we just removed
			{Obj: mf.CRBScheduler},
			{Obj: mf.CRScheduler},
			{Obj: mf.CRBController},
			{Obj: mf.CRController},
			{Obj: mf.Crd},
		}
	}
	// TODO
	return nil
}

func GetManifests(plat platform.Platform) (Manifests, error) {
	mf, err := GetManifestsForNamespace(plat, "")
	if err != nil {
		return mf, err
	}
	ns, err = manifests.Namespace(manifests.ComponentSchedulerPlugin)
	if err != nil {
		return mf, err
	}
	mf.Namespace = ns
	return mf, nil
}

func GetManifestsForNamespace(plat platform.Platform, namespace string) (Manifests, error) {
	var err error
	mf := Manifests{
		plat:      plat,
		namespace: namespace,
	}
	mf.Crd, err = manifests.SchedulerCRD()
	if err != nil {
		return mf, err
	}

	mf.ConfigMap, err = manifests.ConfigMap(manifests.ComponentSchedulerPlugin, "")
	if err != nil {
		return mf, err
	}
	mf.SAScheduler, err = manifests.ServiceAccount(manifests.ComponentSchedulerPlugin, manifests.SubComponentSchedulerPluginScheduler)
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
	mf.DPScheduler, err = manifests.Deployment(manifests.ComponentSchedulerPlugin, manifests.SubComponentSchedulerPluginScheduler)
	if err != nil {
		return mf, err
	}

	mf.SAController, err = manifests.ServiceAccount(manifests.ComponentSchedulerPlugin, manifests.SubComponentSchedulerPluginController)
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
	mf.DPController, err = manifests.Deployment(manifests.ComponentSchedulerPlugin, manifests.SubComponentSchedulerPluginController)
	if err != nil {
		return mf, err
	}

	return mf, nil
}

func newInt32(value int32) *int32 {
	return &value
}
