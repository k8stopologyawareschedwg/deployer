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

package wait

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	machineconfigv1 "github.com/openshift/machine-config-operator/pkg/apis/machineconfiguration.openshift.io/v1"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/ready"
	"github.com/k8stopologyawareschedwg/deployer/pkg/tlog"
)

func PodsToBeRunningByRegex(hp *deployer.Helper, log tlog.Logger, namespace, name string) error {
	log.Printf("wait for all the pods in group %s %s to be running and ready", namespace, name)
	return wait.PollImmediate(1*time.Second, 3*time.Minute, func() (bool, error) {
		pods, err := hp.GetPodsByPattern(namespace, fmt.Sprintf("%s-*", name))
		if err != nil {
			return false, err
		}
		if len(pods) == 0 {
			log.Printf("no pods found for %s %s", namespace, name)
			return false, nil
		}

		for _, pod := range pods {
			if pod.Status.Phase != corev1.PodRunning {
				log.Printf("pod %s %s not ready yet (%s)", pod.Namespace, pod.Name, pod.Status.Phase)
				return false, nil
			}
		}
		log.Printf("all the pods in daemonset %s %s are running and ready!", namespace, name)
		return true, nil
	})
}

func PodsToBeGoneByRegex(hp *deployer.Helper, log tlog.Logger, namespace, name string) error {
	log.Printf("wait for all the pods in deployment %s %s to be gone", namespace, name)
	return wait.PollImmediate(10*time.Second, 3*time.Minute, func() (bool, error) {
		pods, err := hp.GetPodsByPattern(namespace, fmt.Sprintf("%s-*", name))
		if err != nil {
			return false, err
		}
		if len(pods) > 0 {
			return false, fmt.Errorf("still %d pods found for %s %s", len(pods), namespace, name)
		}
		log.Printf("all pods gone for deployment %s %s are gone!", namespace, name)
		return true, nil
	})
}

func NamespaceToBeGone(hp *deployer.Helper, log tlog.Logger, namespace string) error {
	log.Printf("wait for the namespace %q to be gone", namespace)
	return wait.PollImmediate(1*time.Second, 3*time.Minute, func() (bool, error) {
		nsKey := types.NamespacedName{
			Name: namespace,
		}
		ns := corev1.Namespace{} // unused
		err := hp.GetObject(nsKey, &ns)
		if err == nil {
			// still present
			return false, nil
		}
		if !k8serrors.IsNotFound(err) {
			return false, err
		}
		log.Printf("namespace %q gone!", namespace)
		return true, nil
	})
}

func DaemonSetToBeRunning(hp *deployer.Helper, log tlog.Logger, namespace, name string) error {
	log.Printf("wait for the daemonset %q %q to be running", namespace, name)
	return wait.PollImmediate(3*time.Second, 3*time.Minute, func() (bool, error) {
		return hp.IsDaemonSetRunning(namespace, name)
	})
}

func DaemonSetToBeGone(hp *deployer.Helper, log tlog.Logger, namespace, name string) error {
	log.Printf("wait for the daemonset %q %q to be gone", namespace, name)
	return wait.PollImmediate(3*time.Second, 3*time.Minute, func() (bool, error) {
		return hp.IsDaemonSetGone(namespace, name)
	})
}

func MachineConfigPoolToBeUpdated(hp *deployer.Helper, log tlog.Logger, name string, mcpLabels map[string]string) error {
	// we target a single MCP anyway, so let's figure it out once outside the loop to save calls and CPU time
	mcps, err := hp.ListMachineConfigPools()
	if err != nil {
		return err
	}
	mcp, err := findMCPByLabels(mcps, mcpLabels, log)
	if err != nil {
		return err
	}

	log.Printf("wait for the machineconfig pool %q to be updated", mcp.Name)
	return wait.PollImmediate(5*time.Second, 60*time.Minute, func() (bool, error) {
		mcp, err := hp.GetMachineConfigPoolByName(mcp.Name)
		if err != nil {
			return false, err
		}
		return ready.MachineConfigPool(mcp, name), nil
	})
}

func findMCPByLabels(mcps []machineconfigv1.MachineConfigPool, mcpLabels map[string]string, log tlog.Logger) (*machineconfigv1.MachineConfigPool, error) {
	mcpSelector := &metav1.LabelSelector{
		MatchLabels: mcpLabels,
	}

	for i := range mcps {
		mcp := &mcps[i]

		selector, err := metav1.LabelSelectorAsSelector(mcpSelector)
		if err != nil {
			log.Debugf("bad machine config pool selector %q", mcpSelector.String())
			continue
		}

		mcpLabels := labels.Set(mcp.Labels)
		if selector.Matches(mcpLabels) {
			return mcp, nil
		}
	}

	return nil, fmt.Errorf("failed to find MachineConfigPool for the node group with the selector %q", mcpSelector.String())
}
