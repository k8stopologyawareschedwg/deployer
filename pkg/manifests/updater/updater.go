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
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/tlog"
)

const (
	RTE = "RTE"
	NFD = "NFD"
)

const (
	namespaceOCP      = "openshift-monitoring"
	serviceAccountOCP = "node-exporter"
)

type ManifestsHandler interface {
	GetManifests() Manifests
	Update(options UpdateOptions) ManifestsHandler
	ToObjects() []client.Object
	ToCreatableObjects(hp *deployer.Helper, log tlog.Logger) []deployer.WaitableObject
	ToDeletableObjects(hp *deployer.Helper, log tlog.Logger) []deployer.WaitableObject
}

type Manifests struct {
	Namespace          *corev1.Namespace
	ServiceAccount     *corev1.ServiceAccount
	ClusterRole        *rbacv1.ClusterRole
	ClusterRoleBinding *rbacv1.ClusterRoleBinding
	ConfigMap          *corev1.ConfigMap
	DaemonSet          *appsv1.DaemonSet
	// internal fields
	plat           platform.Platform
	namespace      string
	serviceAccount string
}

type UpdateOptions struct {
	ConfigData       string
	PullIfNotPresent bool
}

func GetManifestsHandler(plat platform.Platform, updaterType string) (ManifestsHandler, error) {
	switch updaterType {
	case RTE:
		return rteManifestsHandler(plat)
	case NFD:
		return nfdManifestsHandler(plat)
	}
	return nil, fmt.Errorf("%q is invalid updater type", updaterType)
}
