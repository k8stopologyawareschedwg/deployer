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

package wait

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	k8swait "k8s.io/apimachinery/pkg/util/wait"

	machineconfigv1 "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
)

func (wt Waiter) ForMachineConfigPoolCondition(ctx context.Context, mcp *machineconfigv1.MachineConfigPool, condType machineconfigv1.MachineConfigPoolConditionType) error {
	err := k8swait.Poll(wt.PollInterval, wt.PollTimeout, func() (bool, error) {
		updatedMcp := machineconfigv1.MachineConfigPool{}
		key := ObjectKeyFromObject(mcp)
		err := wt.Cli.Get(ctx, key.AsKey(), &updatedMcp)
		if err != nil {
			return false, err
		}
		for _, cond := range updatedMcp.Status.Conditions {
			if cond.Type == condType {
				if cond.Status == corev1.ConditionTrue {
					return true, nil
				} else {
					wt.Log.Info("mcp: condition type does not have expected status",
						"machineConfigPoolName", updatedMcp.Name,
						"conditionType", cond.Type,
						"current status", cond.Status,
						"expected status", corev1.ConditionTrue,
					)
					return false, nil
				}
			}
		}
		wt.Log.Info("mcp: condition type was not found", "mcp-name", updatedMcp.Name, "conditionType", condType)
		return false, nil
	})
	return err
}
