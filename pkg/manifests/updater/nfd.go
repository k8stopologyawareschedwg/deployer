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

package updater

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/wait"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
	"github.com/k8stopologyawareschedwg/deployer/pkg/tlog"
)

type NFDManifests struct {
	Manifests
}

func (mf NFDManifests) GetManifests() Manifests {
	return mf.Manifests
}

func (mf NFDManifests) clone() NFDManifests {
	ret := NFDManifests{
		Manifests{
			plat:           mf.plat,
			namespace:      mf.namespace,
			serviceAccount: mf.serviceAccount,
			// objects
			ClusterRole:        mf.ClusterRole.DeepCopy(),
			ClusterRoleBinding: mf.ClusterRoleBinding.DeepCopy(),
			DaemonSet:          mf.DaemonSet.DeepCopy(),
		},
	}
	if mf.plat == platform.Kubernetes {
		ret.Namespace = mf.Namespace.DeepCopy()
		ret.ServiceAccount = mf.ServiceAccount.DeepCopy()
	}
	return ret
}

func (mf NFDManifests) Update(options UpdateOptions) ManifestsHandler {
	ret := mf.clone()
	if ret.plat == platform.Kubernetes {
		ret.ServiceAccount.Namespace = mf.namespace
	}
	if len(options.ConfigData) > 0 {
		ret.ConfigMap = createNFDConfigMap(mf.namespace, options.ConfigData)
	}

	ret.DaemonSet.Namespace = mf.namespace
	ret.DaemonSet.Spec.Template.Spec.ServiceAccountName = mf.serviceAccount
	manifests.UpdateClusterRoleBinding(ret.ClusterRoleBinding, mf.serviceAccount, mf.namespace)
	manifests.UpdateNodeFeatureDiscoveryDaemonSet(ret.plat, ret.DaemonSet, ret.ConfigMap, options.PullIfNotPresent)
	return ret
}

func createNFDConfigMap(namespace string, configData string) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{
		// TODO: why is this needed?
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nfd-config",
			Namespace: namespace,
		},
		Data: map[string]string{
			"config.yaml": configData,
		},
	}
	return cm
}

func (mf NFDManifests) ToObjects() []client.Object {
	objs := []client.Object{
		mf.ClusterRole,
		mf.ClusterRoleBinding,
	}
	if mf.ConfigMap != nil {
		objs = append(objs, mf.ConfigMap)
	}
	objs = append(objs, mf.DaemonSet)
	if mf.plat == platform.Kubernetes {
		kubeObjs := []client.Object{
			mf.Namespace,
			mf.ServiceAccount,
		}
		return append(kubeObjs, objs...)
	}
	return objs
}

func (mf NFDManifests) ToCreatableObjects(hp *deployer.Helper, log tlog.Logger) []deployer.WaitableObject {
	objs := []deployer.WaitableObject{
		{Obj: mf.ClusterRole},
		{Obj: mf.ClusterRoleBinding},
		{
			Obj:  mf.DaemonSet,
			Wait: func() error { return wait.PodsToBeRunningByRegex(hp, log, mf.DaemonSet.Namespace, mf.DaemonSet.Name) },
		},
	}
	if mf.ConfigMap != nil {
		objs = append([]deployer.WaitableObject{{Obj: mf.ConfigMap}}, objs...)
	}
	if mf.plat == platform.Kubernetes {
		kubeObjs := []deployer.WaitableObject{
			{Obj: mf.Namespace},
			{Obj: mf.ServiceAccount},
		}
		return append(kubeObjs, objs...)
	}
	return objs
}

func (mf NFDManifests) ToDeletableObjects(hp *deployer.Helper, log tlog.Logger) []deployer.WaitableObject {
	if mf.plat == platform.Kubernetes {
		return []deployer.WaitableObject{
			{
				Obj:  mf.Namespace,
				Wait: func() error { return wait.NamespaceToBeGone(hp, log, mf.Namespace.Name) },
			},
			// no need to remove objects created inside the namespace we just removed
			{Obj: mf.ClusterRole},
			{Obj: mf.ClusterRoleBinding},
			{Obj: mf.ServiceAccount},
		}
	}
	objs := []deployer.WaitableObject{
		{
			Obj:  mf.DaemonSet,
			Wait: func() error { return wait.PodsToBeGoneByRegex(hp, log, mf.DaemonSet.Namespace, mf.DaemonSet.Name) },
		},
		{Obj: mf.ClusterRole},
		{Obj: mf.ClusterRoleBinding},
	}
	if mf.ConfigMap != nil {
		objs = append(objs, deployer.WaitableObject{Obj: mf.ConfigMap})
	}
	return objs
}

func (mf NFDManifests) GetDaemonSet() *appsv1.DaemonSet {
	return mf.DaemonSet
}

func (mf NFDManifests) GetNamespace() *corev1.Namespace {
	return mf.Namespace
}

func nfdManifestsHandler(plat platform.Platform) (NFDManifests, error) {
	var err error
	mf := NFDManifests{
		Manifests{
			plat: plat,
		},
	}
	if plat == platform.Kubernetes {
		mf.Namespace, err = manifests.Namespace(manifests.ComponentNodeFeatureDiscovery)
		if err != nil {
			return mf, err
		}
		mf.namespace = mf.Namespace.Name

		mf.ServiceAccount, err = manifests.ServiceAccount(manifests.ComponentNodeFeatureDiscovery, "")
		if err != nil {
			return mf, err
		}
		mf.serviceAccount = mf.ServiceAccount.Name
	} else {
		mf.namespace = namespaceOCP
		mf.serviceAccount = serviceAccountOCP
	}
	mf.ClusterRole, err = manifests.ClusterRole(manifests.ComponentNodeFeatureDiscovery, "")
	if err != nil {
		return mf, err
	}
	mf.ClusterRoleBinding, err = manifests.ClusterRoleBinding(manifests.ComponentNodeFeatureDiscovery, "")
	if err != nil {
		return mf, err
	}
	mf.DaemonSet, err = manifests.DaemonSet(manifests.ComponentNodeFeatureDiscovery)
	if err != nil {
		return mf, err
	}
	return mf, nil
}
