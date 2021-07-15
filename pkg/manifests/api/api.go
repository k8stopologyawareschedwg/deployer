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
	"github.com/fromanirh/deployer/pkg/manifests"
	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Manifests struct {
	Crd *apiextensionv1.CustomResourceDefinition
}

func (mf Manifests) ToObjects() []runtime.Object {
	return []runtime.Object{
		mf.Crd,
	}
}

func (mf Manifests) UpdateNamespace() Manifests {
	ret := Manifests{
		Crd: mf.Crd.DeepCopy(),
	}
	// nothing to do atm
	return ret
}

func GetManifests() (Manifests, error) {
	var err error
	mf := Manifests{}

	mf.Crd, err = manifests.APICRD()
	if err != nil {
		return mf, err
	}

	return mf, nil
}
