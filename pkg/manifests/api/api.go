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
	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fromanirh/deployer/pkg/deployer"
	"github.com/fromanirh/deployer/pkg/deployer/platform"
	"github.com/fromanirh/deployer/pkg/manifests"
)

type Manifests struct {
	Crd  *apiextensionv1.CustomResourceDefinition
	plat platform.Platform
}

func (mf Manifests) ToObjects() []client.Object {
	return []client.Object{
		mf.Crd,
	}
}

func (mf Manifests) ToCreatableObjects(hp *deployer.Helper, log deployer.Logger) []deployer.WaitableObject {
	return []deployer.WaitableObject{
		deployer.WaitableObject{Obj: mf.Crd},
	}
}

func (mf Manifests) ToDeletableObjects(hp *deployer.Helper, log deployer.Logger) []deployer.WaitableObject {
	return []deployer.WaitableObject{
		deployer.WaitableObject{Obj: mf.Crd},
	}
}

func (mf Manifests) Clone() Manifests {
	return Manifests{
		Crd: mf.Crd.DeepCopy(),
	}
}

func (mf Manifests) Update() Manifests {
	ret := mf.Clone()
	// nothing to do atm
	return ret
}

func GetManifests(plat platform.Platform) (Manifests, error) {
	var err error
	mf := Manifests{
		plat: plat,
	}

	mf.Crd, err = manifests.APICRD()
	if err != nil {
		return mf, err
	}

	return mf, nil
}
