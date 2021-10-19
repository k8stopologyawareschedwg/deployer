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
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/wait"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
	"github.com/k8stopologyawareschedwg/deployer/pkg/tlog"
)

const (
	NamespaceOpenShift      = "openshift-monitoring"
	ServiceAccountOpenShift = "node-exporter"
)

type Manifests struct {
	ServiceAccount *corev1.ServiceAccount
	Role           *rbacv1.Role
	RoleBinding    *rbacv1.RoleBinding
	ConfigMap      *corev1.ConfigMap
	DaemonSet      *appsv1.DaemonSet
	// internal fields
	plat           platform.Platform
	namespace      string
	serviceAccount string
}

func (mf Manifests) Clone() Manifests {
	ret := Manifests{
		plat:           mf.plat,
		namespace:      mf.namespace,
		serviceAccount: mf.serviceAccount,
		// objects
		Role:        mf.Role.DeepCopy(),
		RoleBinding: mf.RoleBinding.DeepCopy(),
		DaemonSet:   mf.DaemonSet.DeepCopy(),
	}
	if mf.plat == platform.Kubernetes {
		ret.ServiceAccount = mf.ServiceAccount.DeepCopy()
	}
	return ret
}

type UpdateOptions struct {
	ConfigData       string
	PullIfNotPresent bool
	Namespace        string
}

func (mf Manifests) Update(options UpdateOptions) Manifests {
	if options.Namespace != "" {
		mf.namespace = options.Namespace
	}

	ret := mf.Clone()
	if ret.plat == platform.Kubernetes {
		ret.ServiceAccount.Namespace = mf.namespace
	}
	if len(options.ConfigData) > 0 {
		ret.ConfigMap = createConfigMap(mf.namespace, options.ConfigData)
	}

	ret.DaemonSet.Namespace = mf.namespace
	ret.DaemonSet.Spec.Template.Spec.ServiceAccountName = mf.serviceAccount
	ret.Role.Namespace = mf.namespace
	manifests.UpdateRoleBinding(ret.RoleBinding, mf.serviceAccount, mf.namespace)
	manifests.UpdateResourceTopologyExporterDaemonSet(ret.plat, ret.DaemonSet, ret.ConfigMap, options.PullIfNotPresent)
	return ret
}

func createConfigMap(namespace string, configData string) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{
		// TODO: why is this needed?
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rte-config",
			Namespace: namespace,
		},
		Data: map[string]string{
			"config.yaml": configData,
		},
	}
	return cm
}

func (mf Manifests) ToObjects() []client.Object {
	var objs []client.Object
	if mf.ServiceAccount != nil {
		objs = append(objs, mf.ServiceAccount)
	}
	if mf.ConfigMap != nil {
		objs = append(objs, mf.ConfigMap)
	}
	return append(objs,
		mf.Role,
		mf.RoleBinding,
		mf.DaemonSet,
	)
}

func (mf Manifests) ToCreatableObjects(hp *deployer.Helper, log tlog.Logger) []deployer.WaitableObject {
	var objs []deployer.WaitableObject
	if mf.ServiceAccount != nil {
		objs = append(objs, deployer.WaitableObject{
			Obj: mf.ServiceAccount,
		})
	}
	if mf.ConfigMap != nil {
		objs = append(objs, deployer.WaitableObject{
			Obj: mf.ConfigMap,
		})
	}
	return append(objs,
		deployer.WaitableObject{Obj: mf.Role},
		deployer.WaitableObject{Obj: mf.RoleBinding},
		deployer.WaitableObject{
			Obj:  mf.DaemonSet,
			Wait: func() error { return wait.DaemonSetToBeRunning(hp, log, mf.DaemonSet.Namespace, mf.DaemonSet.Name) },
		},
	)
}

func (mf Manifests) ToDeletableObjects(hp *deployer.Helper, log tlog.Logger) []deployer.WaitableObject {
	objs := []deployer.WaitableObject{
		{
			Obj:  mf.DaemonSet,
			Wait: func() error { return wait.DaemonSetToBeGone(hp, log, mf.DaemonSet.Namespace, mf.DaemonSet.Name) },
		},
		{Obj: mf.RoleBinding},
		{Obj: mf.Role},
	}
	if mf.ConfigMap != nil {
		objs = append(objs, deployer.WaitableObject{Obj: mf.ConfigMap})
	}
	if mf.ServiceAccount != nil {
		objs = append(objs, deployer.WaitableObject{
			Obj: mf.ServiceAccount,
		})
	}
	return objs
}

func GetManifests(plat platform.Platform) (Manifests, error) {
	var err error
	mf := Manifests{
		plat: plat,
	}
	if plat == platform.Kubernetes {
		mf.ServiceAccount, err = manifests.ServiceAccount(manifests.ComponentResourceTopologyExporter, "")
		if err != nil {
			return mf, err
		}
		mf.serviceAccount = mf.ServiceAccount.Name
	} else {
		mf.serviceAccount = ServiceAccountOpenShift
	}
	mf.Role, err = manifests.Role(manifests.ComponentResourceTopologyExporter, "")
	if err != nil {
		return mf, err
	}
	mf.RoleBinding, err = manifests.RoleBinding(manifests.ComponentResourceTopologyExporter, "")
	if err != nil {
		return mf, err
	}
	mf.DaemonSet, err = manifests.DaemonSet(manifests.ComponentResourceTopologyExporter)
	if err != nil {
		return mf, err
	}
	return mf, nil
}

type ExistingManifests struct {
	Existing            Manifests
	ServiceAccountError error
	RoleError           error
	RoleBindingError    error
	ConfigMapError      error
	DaemonSetError      error
}

func (mf Manifests) FromClient(ctx context.Context, cli client.Client) ExistingManifests {
	ret := ExistingManifests{
		Existing: Manifests{
			plat:           mf.plat,
			namespace:      mf.namespace,
			serviceAccount: mf.serviceAccount,
		},
	}

	ro := rbacv1.Role{}
	if ret.RoleError = cli.Get(ctx, client.ObjectKeyFromObject(mf.Role), &ro); ret.RoleError == nil {
		ret.Existing.Role = &ro
	}
	rb := rbacv1.RoleBinding{}
	if ret.RoleBindingError = cli.Get(ctx, client.ObjectKeyFromObject(mf.RoleBinding), &rb); ret.RoleBindingError == nil {
		ret.Existing.RoleBinding = &rb
	}
	ds := appsv1.DaemonSet{}
	if ret.DaemonSetError = cli.Get(ctx, client.ObjectKeyFromObject(mf.DaemonSet), &ds); ret.DaemonSetError == nil {
		ret.Existing.DaemonSet = &ds
	}
	if mf.ServiceAccount != nil {
		sa := corev1.ServiceAccount{}
		if ret.ServiceAccountError = cli.Get(ctx, client.ObjectKeyFromObject(mf.ServiceAccount), &sa); ret.ServiceAccountError == nil {
			ret.Existing.ServiceAccount = &sa
		}
	}
	if mf.ConfigMap != nil {
		cm := corev1.ConfigMap{}
		if ret.ConfigMapError = cli.Get(ctx, client.ObjectKeyFromObject(mf.ConfigMap), &cm); ret.ConfigMapError == nil {
			ret.Existing.ConfigMap = &cm
		}
	}
	return ret
}
