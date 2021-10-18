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

package api

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	apimanifests "github.com/k8stopologyawareschedwg/deployer/pkg/manifests/api"
	"github.com/k8stopologyawareschedwg/deployer/pkg/tlog"
)

type Options struct {
	Platform platform.Platform
}

func SetupNamespace(plat platform.Platform) (*corev1.Namespace, string, error) {
	return nil, "", fmt.Errorf("the API is a cluster scoped resource")
}

func Deploy(log tlog.Logger, opts Options) error {
	var err error
	log.Printf("deploying topology-aware-scheduling API...")

	mf, err := apimanifests.GetManifests(opts.Platform)
	if err != nil {
		return err
	}
	log.Debugf("API manifests loaded")

	hp, err := deployer.NewHelper("API", log)
	if err != nil {
		return err
	}

	if err = hp.CreateObject(mf.Crd); err != nil {
		return err
	}

	log.Printf("...deployed topology-aware-scheduling API!")
	return nil
}

func Remove(log tlog.Logger, opts Options) error {
	var err error
	log.Printf("removing topology-aware-scheduling API...")

	mf, err := apimanifests.GetManifests(opts.Platform)
	if err != nil {
		return err
	}
	log.Debugf("API manifests loaded")

	hp, err := deployer.NewHelper("API", log)
	if err != nil {
		return err
	}

	if err = hp.DeleteObject(mf.Crd); err != nil {
		return err
	}

	log.Printf("...removed topology-aware-scheduling API!")
	return nil
}
