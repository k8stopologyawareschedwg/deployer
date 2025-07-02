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

	"github.com/go-logr/logr"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/wait"
	rtemf "github.com/k8stopologyawareschedwg/deployer/pkg/manifests/rte"
	"github.com/k8stopologyawareschedwg/deployer/pkg/objectwait"
)

func Creatable(mf rtemf.Manifests, cli client.Client, log logr.Logger) []objectwait.WaitableObject {
	var objs []objectwait.WaitableObject
	if mf.ConfigMap != nil {
		objs = append(objs, objectwait.WaitableObject{
			Obj: mf.ConfigMap,
		})
	}

	if mf.SecurityContextConstraint != nil {
		objs = append(objs, objectwait.WaitableObject{
			Obj: mf.SecurityContextConstraint,
		})
	}

	if mf.SecurityContextConstraintV2 != nil {
		objs = append(objs, objectwait.WaitableObject{
			Obj: mf.SecurityContextConstraintV2,
		})
	}

	if mf.MachineConfig != nil {
		// TODO: we should add functionality to wait for the MCP update
		objs = append(objs, objectwait.WaitableObject{
			Obj: mf.MachineConfig,
		})
	}

	key := wait.ObjectKey{
		Namespace: mf.DaemonSet.Namespace,
		Name:      mf.DaemonSet.Name,
	}

	return append(objs,
		objectwait.WaitableObject{Obj: mf.Role},
		objectwait.WaitableObject{Obj: mf.RoleBinding},
		objectwait.WaitableObject{Obj: mf.ClusterRole},
		objectwait.WaitableObject{Obj: mf.ClusterRoleBinding},
		objectwait.WaitableObject{Obj: mf.ServiceAccount},
		objectwait.WaitableObject{Obj: mf.DefaultNetworkPolicy},
		objectwait.WaitableObject{Obj: mf.APIServerNetworkPolicy},
		objectwait.WaitableObject{Obj: mf.MetricsServerNetworkPolicy},
		objectwait.WaitableObject{
			Obj: mf.DaemonSet,
			Wait: func(ctx context.Context) error {
				_, err := wait.With(cli, log).ForDaemonSetReadyByKey(ctx, key)
				return err
			},
		},
	)
}

func Deletable(mf rtemf.Manifests, cli client.Client, log logr.Logger) []objectwait.WaitableObject {
	objs := []objectwait.WaitableObject{
		{
			Obj: mf.DaemonSet,
			Wait: func(ctx context.Context) error {
				return wait.With(cli, log).ForDaemonSetDeleted(ctx, mf.DaemonSet.Namespace, mf.DaemonSet.Name)
			},
		},
		{Obj: mf.Role},
		{Obj: mf.RoleBinding},
		{Obj: mf.ClusterRole},
		{Obj: mf.ClusterRoleBinding},
		{Obj: mf.ServiceAccount},
		{Obj: mf.DefaultNetworkPolicy},
		{Obj: mf.APIServerNetworkPolicy},
		{Obj: mf.MetricsServerNetworkPolicy},
	}
	if mf.ConfigMap != nil {
		objs = append(objs, objectwait.WaitableObject{Obj: mf.ConfigMap})
	}
	if mf.SecurityContextConstraint != nil {
		objs = append(objs, objectwait.WaitableObject{
			Obj: mf.SecurityContextConstraint,
		})
	}
	if mf.SecurityContextConstraintV2 != nil {
		objs = append(objs, objectwait.WaitableObject{
			Obj: mf.SecurityContextConstraintV2,
		})
	}
	if mf.MachineConfig != nil {
		objs = append(objs, objectwait.WaitableObject{
			// TODO: we should add functionality to wait for the MCP update
			Obj: mf.MachineConfig,
		})
	}
	return objs
}
