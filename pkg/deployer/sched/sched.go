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
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/fromanirh/deployer/pkg/deployer"
	schedmanifests "github.com/fromanirh/deployer/pkg/manifests/sched"
)

type Options struct{}

func Deploy(logger *log.Logger, opts Options) error {
	var err error

	mf, err := schedmanifests.GetManifests()
	if err != nil {
		return err
	}

	mf = mf.UpdateNamespace()
	logger.Printf("SCHED manifests loaded")

	hp, err := deployer.NewHelper("SCHED")
	if err != nil {
		return err
	}

	if err = hp.CreateObject(mf.Namespace); err != nil {
		return err
	}
	if err = hp.CreateObject(mf.ServiceAccount); err != nil {
		return err
	}
	if err = hp.CreateObject(mf.ClusterRole); err != nil {
		return err
	}
	if err = hp.CreateObject(mf.CRBKubernetesScheduler); err != nil {
		return err
	}
	if err = hp.CreateObject(mf.CRBNodeResourceTopology); err != nil {
		return err
	}
	if err = hp.CreateObject(mf.CRBVolumeScheduler); err != nil {
		return err
	}
	if err = hp.CreateObject(mf.RoleBinding); err != nil {
		return err
	}
	if err = hp.CreateObject(mf.ConfigMap); err != nil {
		return err
	}

	dp := manifests.UpdateSchedulerPluginDeployment(mf.Deployment)
	if err = hp.CreateObject(dp); err != nil {
		return err
	}

	return nil
}

func Remove(logger *log.Logger, opts Options) error {
	var err error

	mf, err := schedmanifests.GetManifests()
	if err != nil {
		return err
	}

	mf = mf.UpdateNamespace()
	logger.Printf("SCHED manifests loaded")

	hp, err := deployer.NewHelper("SCHED")
	if err != nil {
		return err
	}

	if err = hp.DeleteObject(mf.Namespace); err != nil {
		return err
	}

	nsKey := types.NamespacedName{
		Name:      mf.Namespace.Name,
		Namespace: metav1.NamespaceNone,
	}

	err = hp.WaitForObjectToBeDeleted(nsKey, &corev1.Namespace{})
	if err != nil {
		return err
	}
	if err = hp.DeleteObject(mf.RoleBinding); err != nil {
		return err
	}
	if err = hp.DeleteObject(mf.CRBKubernetesScheduler); err != nil {
		return err
	}
	if err = hp.DeleteObject(mf.CRBNodeResourceTopology); err != nil {
		return err
	}
	if err = hp.DeleteObject(mf.CRBVolumeScheduler); err != nil {
		return err
	}
	if err = hp.DeleteObject(mf.ClusterRole); err != nil {
		return err
	}
	return nil
}
