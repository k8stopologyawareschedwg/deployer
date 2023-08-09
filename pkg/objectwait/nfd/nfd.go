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
 * Copyright 2022 Red Hat, Inc.
 */

package nfd

import (
	"context"

	"github.com/go-logr/logr"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/wait"
	nfdmf "github.com/k8stopologyawareschedwg/deployer/pkg/manifests/nfd"
	"github.com/k8stopologyawareschedwg/deployer/pkg/objectwait"
)

func Creatable(mf nfdmf.Manifests, cli client.Client, log logr.Logger) []objectwait.WaitableObject {
	return []objectwait.WaitableObject{
		{Obj: mf.SATopologyUpdater},
		{Obj: mf.CRTopologyUpdater},
		{Obj: mf.CRBTopologyUpdater},
		{
			Obj: mf.DSTopologyUpdater,
			Wait: func(ctx context.Context) error {
				_, err := wait.With(cli, log).ForDaemonSetReady(ctx, mf.DSTopologyUpdater)
				return err
			},
		},
	}
}

func Deletable(mf nfdmf.Manifests, cli client.Client, log logr.Logger) []objectwait.WaitableObject {
	return []objectwait.WaitableObject{
		{Obj: mf.CRBTopologyUpdater},
		{Obj: mf.CRTopologyUpdater},
	}
}
