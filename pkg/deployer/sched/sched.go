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
	"github.com/fromanirh/deployer/pkg/deployer"
	schedmanifests "github.com/fromanirh/deployer/pkg/manifests/sched"
)

type Options struct{}

func Deploy(log deployer.Logger, opts Options) error {
	var err error
	log.Printf("deploying topology-aware-scheduling scheduler plugin...")

	mf, err := schedmanifests.GetManifests()
	if err != nil {
		return err
	}

	mf = mf.UpdateNamespace().UpdatePullspecs()
	log.Debugf("SCD manifests loaded")

	hp, err := deployer.NewHelper("SCD", log)
	if err != nil {
		return err
	}

	for _, obj := range mf.ToObjects() {
		if err := hp.CreateObject(obj); err != nil {
			return err
		}
	}
	// TODO: wait for the deployment to be running

	log.Printf("...deployed topology-aware-scheduling scheduler plugin!")
	return nil
}

func Remove(log deployer.Logger, opts Options) error {
	var err error
	log.Printf("removing topology-aware-scheduling scheduler plugin...")

	mf, err := schedmanifests.GetManifests()
	if err != nil {
		return err
	}

	mf = mf.UpdateNamespace()
	log.Debugf("SCD manifests loaded")

	hp, err := deployer.NewHelper("SCD", log)
	if err != nil {
		return err
	}

	if err = hp.DeleteObject(mf.Deployment); err != nil {
		return err
	}
	// TODO: wait for the deployment to be gone
	if err = hp.DeleteObject(mf.ConfigMap); err != nil {
		return err
	}

	if err = hp.DeleteObject(mf.CRBKubernetesScheduler); err != nil {
		return err
	}
	if err = hp.DeleteObject(mf.CRBVolumeScheduler); err != nil {
		return err
	}
	if err = hp.DeleteObject(mf.CRBNodeResourceTopology); err != nil {
		return err
	}
	if err = hp.DeleteObject(mf.ClusterRole); err != nil {
		return err
	}
	if err = hp.DeleteObject(mf.RoleBinding); err != nil {
		return err
	}
	if err = hp.DeleteObject(mf.ServiceAccount); err != nil {
		return err
	}

	log.Printf("...removed topology-aware-scheduling scheduler plugin!")
	return nil
}
