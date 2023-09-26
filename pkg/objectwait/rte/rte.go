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
	"fmt"
	"time"

	"github.com/go-logr/logr"

	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/wait"
	rtemf "github.com/k8stopologyawareschedwg/deployer/pkg/manifests/rte"
	"github.com/k8stopologyawareschedwg/deployer/pkg/multierrors"
	"github.com/k8stopologyawareschedwg/deployer/pkg/objectwait"

	mcov1 "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
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

	if mf.MachineConfig != nil {
		objs = append(objs, objectwait.WaitableObject{
			Obj: mf.MachineConfig,
			Wait: func(ctx context.Context) error {
				mcps, err := getMPCsForMC(cli, ctx, *mf.MachineConfig)
				if err != nil {
					return err
				}
				pollInterval := 30 * time.Second
				pollTimeout := 30 * time.Minute
				err = waitForMachineConfigPoolsCondition(cli, log, ctx, mcps, mcov1.MachineConfigPoolUpdating, pollInterval, pollTimeout)
				if err != nil {
					return err
				}
				return waitForMachineConfigPoolsCondition(cli, log, ctx, mcps, mcov1.MachineConfigPoolUpdated, pollInterval, pollTimeout)
			},
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
	}
	if mf.ConfigMap != nil {
		objs = append(objs, objectwait.WaitableObject{Obj: mf.ConfigMap})
	}
	if mf.SecurityContextConstraint != nil {
		objs = append(objs, objectwait.WaitableObject{
			Obj: mf.SecurityContextConstraint,
		})
	}
	if mf.MachineConfig != nil {
		objs = append(objs, objectwait.WaitableObject{
			Obj: mf.MachineConfig,
			Wait: func(ctx context.Context) error {
				mcps, err := getMPCsForMC(cli, ctx, *mf.MachineConfig)
				if err != nil {
					return err
				}
				pollInterval := 30 * time.Second
				pollTimeout := 30 * time.Minute
				err = waitForMachineConfigPoolsCondition(cli, log, ctx, mcps, mcov1.MachineConfigPoolUpdating, pollInterval, pollInterval)
				if err != nil {
					return err
				}
				return waitForMachineConfigPoolsCondition(cli, log, ctx, mcps, mcov1.MachineConfigPoolUpdated, pollInterval, pollTimeout)
			},
		})
	}
	return objs
}

func waitForMachineConfigPoolsCondition(cli client.Client, log logr.Logger, ctx context.Context, mcps []*mcov1.MachineConfigPool, cond mcov1.MachineConfigPoolConditionType, interval, timeout time.Duration) error {
	c1 := make(chan error)
	for _, mcp := range mcps {
		go func(mcp *mcov1.MachineConfigPool) {
			err := wait.With(cli, log).
				Interval(interval).
				Timeout(timeout).
				ForMachineConfigPoolCondition(ctx, mcp, cond)
			c1 <- err
		}(mcp)
	}
	var errs multierrors.MultiErrors
	for idx := 0; idx < len(mcps); idx++ {
		err := <-c1
		errs.Add(err)
	}
	if errs.IsEmpty() {
		return nil
	}

	return fmt.Errorf("problems found while waiting for some MCPs to reach condition %s. %w", cond, &errs)
}

func getMPCsForMC(cli client.Client, ctx context.Context, mc mcov1.MachineConfig) ([]*mcov1.MachineConfigPool, error) {
	mcLabels := mc.GetLabels()
	mcpList := &mcov1.MachineConfigPoolList{}
	if err := cli.List(ctx, mcpList); err != nil {
		return nil, err
	}
	mcps := []*mcov1.MachineConfigPool{}
	for _, mcp := range mcpList.Items {
		selector, err := metav1.LabelSelectorAsSelector(mcp.Spec.MachineConfigSelector)
		if err != nil {
			return nil, err
		}
		if selector.Matches(labels.Set(mcLabels)) {
			mcp := mcp
			mcps = append(mcps, &mcp)
		}
	}
	return mcps, nil
}
