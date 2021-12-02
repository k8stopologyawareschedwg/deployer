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

package ready

import (
	appsv1 "k8s.io/api/apps/v1"

	machineconfigv1 "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"
)

func DaemonSet(ds *appsv1.DaemonSet) bool {
	return (ds.Status.DesiredNumberScheduled > 0 && ds.Status.DesiredNumberScheduled == ds.Status.NumberReady)
}

func MachineConfigPool(mcp *machineconfigv1.MachineConfigPool, mcName string) bool {
	existing := false
	for _, s := range mcp.Status.Configuration.Source {
		if s.Name == mcName {
			existing = true
			break
		}
	}
	// the Machine Config Pool still did not apply the machine config wait for one minute
	if !existing || machineconfigv1.IsMachineConfigPoolConditionFalse(mcp.Status.Conditions, machineconfigv1.MachineConfigPoolUpdated) {
		return false
	}
	return true
}
