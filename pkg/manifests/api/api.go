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
	"context"

	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests/compare"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests/merge"
	"github.com/k8stopologyawareschedwg/deployer/pkg/tlog"
)

type Manifests struct {
	Crd *apiextensionv1.CustomResourceDefinition
	// internal fields
	plat platform.Platform
}

func (mf Manifests) ToObjects() []client.Object {
	return []client.Object{
		mf.Crd,
	}
}

func (mf Manifests) ToCreatableObjects(hp *deployer.Helper, log tlog.Logger) []deployer.WaitableObject {
	return []deployer.WaitableObject{
		{Obj: mf.Crd},
	}
}

func (mf Manifests) ToDeletableObjects(hp *deployer.Helper, log tlog.Logger) []deployer.WaitableObject {
	return []deployer.WaitableObject{
		{Obj: mf.Crd},
	}
}

func (mf Manifests) Clone() Manifests {
	return Manifests{
		plat: mf.plat,
		// objects
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

type ExistingManifests struct {
	Existing Manifests
	CrdError error
}

func (em ExistingManifests) State(mf Manifests) []manifests.ObjectState {
	return []manifests.ObjectState{
		{
			Existing: em.Existing.Crd,
			Error:    em.CrdError,
			Desired:  mf.Crd.DeepCopy(),
			Compare:  compare.Object,
			Merge:    merge.MetadataForUpdate,
		},
	}
}

func (mf Manifests) FromClient(ctx context.Context, cli client.Client) ExistingManifests {
	ret := ExistingManifests{
		Existing: Manifests{
			plat: mf.plat,
		},
	}
	crd := apiextensionv1.CustomResourceDefinition{}
	if ret.CrdError = cli.Get(ctx, client.ObjectKeyFromObject(mf.Crd), &crd); ret.CrdError == nil {
		ret.Existing.Crd = &crd
	}
	return ret
}
