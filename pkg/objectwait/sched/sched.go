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
	"context"

	"github.com/go-logr/logr"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/wait"
	schedmf "github.com/k8stopologyawareschedwg/deployer/pkg/manifests/sched"
	"github.com/k8stopologyawareschedwg/deployer/pkg/objectwait"
)

func Creatable(mf schedmf.Manifests, cli client.Client, log logr.Logger) []objectwait.WaitableObject {
	return []objectwait.WaitableObject{
		{Obj: mf.Crd},
		{Obj: mf.Namespace},
		{Obj: mf.SAScheduler},
		{Obj: mf.CRScheduler},
		{Obj: mf.CRBScheduler},
		{Obj: mf.NPDefaultScheduler},
		{Obj: mf.NPApiServerScheduler},
		{Obj: mf.RSchedulerElect},
		{Obj: mf.RBSchedulerElect},
		{Obj: mf.RBSchedulerAuth},
		{Obj: mf.ConfigMap},
		{
			Obj: mf.DPScheduler,
			Wait: func(ctx context.Context) error {
				_, err := wait.With(cli, log).ForDeploymentComplete(ctx, mf.DPScheduler)
				return err
			},
		},
		{Obj: mf.SAController},
		{Obj: mf.CRController},
		{Obj: mf.CRBController},
		{Obj: mf.RBController},
		{Obj: mf.NPDefaultController},
		{Obj: mf.NPApiServerController},
		{
			Obj: mf.DPController,
			Wait: func(ctx context.Context) error {
				_, err := wait.With(cli, log).ForDeploymentComplete(ctx, mf.DPController)
				return err
			},
		},
	}
}

func Deletable(mf schedmf.Manifests, cli client.Client, log logr.Logger) []objectwait.WaitableObject {
	return []objectwait.WaitableObject{
		{
			Obj: mf.Namespace,
			Wait: func(ctx context.Context) error {
				return wait.With(cli, log).ForNamespaceDeleted(ctx, mf.Namespace.Name)
			},
		},
		// no need to remove objects created inside the namespace we just removed
		{Obj: mf.CRBScheduler},
		{Obj: mf.CRScheduler},
		{Obj: mf.NPDefaultScheduler},
		{Obj: mf.NPApiServerScheduler},
		{Obj: mf.RBSchedulerAuth},
		{Obj: mf.RBSchedulerElect},
		{Obj: mf.RSchedulerElect},
		{Obj: mf.CRBController},
		{Obj: mf.CRController},
		{Obj: mf.RBController},
		{Obj: mf.NPDefaultController},
		{Obj: mf.NPApiServerController},
		{Obj: mf.Crd},
	}
}
