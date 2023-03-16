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

package nodes

import (
	"context"
	"fmt"

	"github.com/onsi/ginkgo/v2"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/client-go/kubernetes"
)

const (
	// RoleWorker contains the worker role
	RoleWorker = "worker"
)

const (
	// LabelRole contains the key for the role label
	LabelRole = "node-role.kubernetes.io"
	// LabelHostname contains the key for the hostname label
	LabelHostname = "kubernetes.io/hostname"
)

// GetWorkerNodes returns all nodes labeled as worker
func GetWorkerNodes(cli *kubernetes.Clientset) ([]corev1.Node, error) {
	return GetNodesByRole(cli, RoleWorker)
}

// GetByRole returns all nodes with the specified role
func GetNodesByRole(cli *kubernetes.Clientset, role string) ([]corev1.Node, error) {
	selector, err := labels.Parse(fmt.Sprintf("%s/%s=", LabelRole, role))
	if err != nil {
		return nil, err
	}
	return GetNodesBySelector(cli, selector)
}

// GetBySelector returns all nodes with the specified selector
func GetNodesBySelector(cli *kubernetes.Clientset, selector labels.Selector) ([]corev1.Node, error) {
	nodes, err := cli.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, err
	}
	return nodes.Items, nil
}

// FilterNodesWithEnoughCores returns all nodes with at least the amount of given CPU allocatable
func FilterNodesWithEnoughCores(nodes []corev1.Node, cpuAmount string) ([]corev1.Node, error) {
	requestCpu := resource.MustParse(cpuAmount)
	fmt.Fprintf(ginkgo.GinkgoWriter, "checking request %v on %d nodes\n", requestCpu, len(nodes))

	resNodes := []corev1.Node{}
	for _, node := range nodes {
		availCpu, ok := node.Status.Allocatable[corev1.ResourceCPU]
		if !ok || availCpu.IsZero() {
			return nil, fmt.Errorf("node %q has no allocatable CPU", node.Name)
		}

		if availCpu.Cmp(requestCpu) < 1 {
			fmt.Fprintf(ginkgo.GinkgoWriter, "node %q available cpu %v requested cpu %v\n", node.Name, availCpu, requestCpu)
			continue
		}

		fmt.Fprintf(ginkgo.GinkgoWriter, "node %q has enough resources, cluster OK\n", node.Name)
		resNodes = append(resNodes, node)
	}

	return resNodes, nil
}
