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
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/fromanirh/deployer/pkg/clientutil"
	"github.com/fromanirh/deployer/pkg/manifests"
)

type Options struct{}

type Manifests struct {
	Namespace          *corev1.Namespace
	ServiceAccount     *corev1.ServiceAccount
	ClusterRole        *rbacv1.ClusterRole
	ClusterRoleBinding *rbacv1.ClusterRoleBinding
	DaemonSet          *appsv1.DaemonSet
}

func (mf Manifests) ToObjects() []runtime.Object {
	return []runtime.Object{
		mf.Namespace,
		mf.ServiceAccount,
		mf.ClusterRole,
		mf.ClusterRoleBinding,
		mf.DaemonSet,
	}
}

func GetManifests() (Manifests, error) {
	var err error
	mf := Manifests{}
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

func Deploy(opts Options) error {
	var err error

	cli, err := clientutil.New()
	if err != nil {
		return err
	}

	mf, err := GetManifests()
	if err != nil {
		return err
	}

	if err := cli.Create(context.TODO(), mf.Namespace); err != nil {
		return err
	}

	if err := cli.Create(context.TODO(), mf.ServiceAccount); err != nil {
		return err
	}

	if err := cli.Create(context.TODO(), mf.ClusterRole); err != nil {
		return err
	}

	if err := cli.Create(context.TODO(), mf.ClusterRoleBinding); err != nil {
		return err
	}

	ds := manifests.UpdateResourceTopologyExporterDaemonSet(mf.DaemonSet)
	if err := cli.Create(context.TODO(), ds); err != nil {
		return err
	}

	// TODO: wait for the DS to go running

	return nil
}

func Remove(opts Options) error {
	return nil
}