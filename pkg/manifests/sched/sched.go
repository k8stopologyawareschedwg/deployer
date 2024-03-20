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
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-logr/logr"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
	rbacupdate "github.com/k8stopologyawareschedwg/deployer/pkg/objectupdate/rbac"
	schedupdate "github.com/k8stopologyawareschedwg/deployer/pkg/objectupdate/sched"
	"github.com/k8stopologyawareschedwg/deployer/pkg/options"
)

const (
	DefaultProfileName         = "topology-aware-scheduler"
	DefaultResyncPeriod        = 5 * time.Second
	DefaultVerbose             = 4
	DefaultCtrlPlaneAffinity   = true
	DefaultLeaderElectResource = manifests.LeaderElectionDefaultNamespace + "/" + manifests.LeaderElectionDefaultName
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

func (mf Manifests) Render(logger logr.Logger, opts options.Scheduler) (Manifests, error) {
	ret := mf.Clone()
	replicas := opts.Replicas
	if replicas <= 0 {
		return ret, fmt.Errorf("non-positive replicas: %d", replicas)
	}
	ret.DPScheduler.Spec.Replicas = newInt32(replicas)
	ret.DPController.Spec.Replicas = newInt32(replicas)

	var err error
	params := manifests.ConfigParams{
		ProfileName: opts.ProfileName,
		Cache:       manifests.NewConfigCacheParams(),
	}

	if len(opts.CacheParamsConfigData) > 0 {
		err = yaml.Unmarshal([]byte(opts.CacheParamsConfigData), params.Cache)
		if err != nil {
			return ret, err
		}
	}

	// always override
	params.Cache.ResyncPeriodSeconds = newInt64(int64(opts.CacheResyncPeriod.Seconds()))

	if len(opts.ScoringStratConfigData) > 0 {
		params.ScoringStrategy = &manifests.ScoringStrategyParams{}
		err = yaml.Unmarshal([]byte(opts.ScoringStratConfigData), params.ScoringStrategy)
		if err != nil {
			return ret, err
		}
	}

	err = schedupdate.SchedulerConfig(ret.ConfigMap, DefaultProfileName, &params)
	if err != nil {
		return ret, err
	}

	schedupdate.SchedulerDeployment(ret.DPScheduler, opts.PullIfNotPresent, opts.CtrlPlaneAffinity, opts.Verbose)
	schedupdate.ControllerDeployment(ret.DPController, opts.PullIfNotPresent, opts.CtrlPlaneAffinity)
	if mf.plat == platform.OpenShift {
		ret.Namespace.Name = NamespaceOpenShift
	}

	ret.SAController.Namespace = ret.Namespace.Name
	rbacupdate.ClusterRoleBinding(ret.CRBController, ret.SAController.Name, ret.Namespace.Name)
	rbacupdate.RoleBinding(ret.RBController, ret.SAController.Name, ret.Namespace.Name)
	ret.DPController.Namespace = ret.Namespace.Name

	ret.SAScheduler.Namespace = ret.Namespace.Name
	rbacupdate.ClusterRoleBinding(ret.CRBScheduler, ret.SAScheduler.Name, ret.Namespace.Name)
	rbacupdate.RoleBinding(ret.RBScheduler, ret.SAScheduler.Name, ret.Namespace.Name)
	ret.DPScheduler.Namespace = ret.Namespace.Name
	ret.ConfigMap.Namespace = ret.Namespace.Name

	return ret, nil
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

func newInt64(value int64) *int64 {
	return &value
}

func toJSON(v any) string {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("<err=%v>", err)
	}
	return string(data)
}
