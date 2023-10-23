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
 * Copyright 2023 Red Hat, Inc.
 */

package api

import (
	"context"

	"github.com/go-logr/logr"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/wait"
	apimf "github.com/k8stopologyawareschedwg/deployer/pkg/manifests/api"
	"github.com/k8stopologyawareschedwg/deployer/pkg/objectwait"
)

func Creatable(mf apimf.Manifests, cli client.Client, log logr.Logger) []objectwait.WaitableObject {
	return []objectwait.WaitableObject{
		{
			Obj: mf.Crd,
			Wait: func(ctx context.Context) error {
				_, err := wait.With(cli, log).ForCRDCreated(ctx, mf.Crd.Name)
				return err
			},
		},
	}
}

func Deletable(mf apimf.Manifests, cli client.Client, log logr.Logger) []objectwait.WaitableObject {
	return []objectwait.WaitableObject{
		{
			Obj: mf.Crd,
			Wait: func(ctx context.Context) error {
				return wait.With(cli, log).ForCRDDeleted(ctx, mf.Crd.Name)
			},
		},
	}
}
