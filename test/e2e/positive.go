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

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha1"
	topologyclientset "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/generated/clientset/versioned"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"github.com/fromanirh/deployer/pkg/clientutil"
	"github.com/fromanirh/deployer/pkg/clientutil/nodes"
	"github.com/fromanirh/deployer/pkg/manifests/rte"
	"github.com/fromanirh/deployer/pkg/manifests/sched"
)

const (
	// TODO: this should be autodetected
	exportedNs = "default"
)

var _ = ginkgo.Describe("[PositiveFlow] Deployer validation", func() {
	ginkgo.Context("with cluster with the expected settings", func() {
		ginkgo.It("it should pass the validation", func() {
			cmdline := []string{
				filepath.Join(binariesPath, "deployer"),
				"validate",
				"--json",
			}
			fmt.Fprintf(ginkgo.GinkgoWriter, "running: %v\n", cmdline)

			cmd := exec.Command(cmdline[0], cmdline[1:]...)
			cmd.Stderr = ginkgo.GinkgoWriter

			out, err := cmd.Output()
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			vo := validationOutput{}
			if err := json.Unmarshal(out, &vo); err != nil {
				ginkgo.Fail(fmt.Sprintf("Error unmarshalling output %q: %v", out, err))
			}
			gomega.Expect(vo.Success).To(gomega.BeTrue())
			gomega.Expect(vo.Errors).To(gomega.BeEmpty())
		})
	})
})

var _ = ginkgo.Describe("[PositiveFlow] Deployer execution", func() {
	ginkgo.Context("with a running cluster without any components", func() {
		ginkgo.BeforeEach(func() {
			err := deploy()
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		})

		ginkgo.AfterEach(func() {
			err := remove()
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		})

		ginkgo.It("should perform overall deployment and verify all pods are running", func() {
			ginkgo.By("checking that resource-topology-exporter pod is running")
			mf, err := rte.GetManifests()
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			mf = mf.UpdateNamespace()
			waitPodsToBeRunningByRegex(fmt.Sprintf("%s-*", mf.DaemonSet.Name))

			ginkgo.By("checking that topo-aware-scheduler pod is running")
			mfs, err := sched.GetManifests()
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			mfs = mfs.UpdateNamespace()
			waitPodsToBeRunningByRegex(fmt.Sprintf("%s-*", mfs.Deployment.Name))

			ginkgo.By("checking that noderesourcetopolgy has some information in it")
			tc, err := clientutil.NewTopologyClient()
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			workers, err := nodes.GetWorkers()
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			for _, node := range workers {
				ginkgo.By(fmt.Sprintf("checking node resource topology for %q", node.Name))

				// the name of the nrt object is the same as the worker node's name
				nrt := getNodeResourceTopology(tc, exportedNs, node.Name)
				// we check CPUs because that's the only resource we know it will always be available
				hasCPU := false
				for _, zone := range nrt.Zones {
					for _, resource := range zone.Resources {
						if resource.Name == string(corev1.ResourceCPU) && resource.Capacity.Size() >= 1 {
							hasCPU = true
						}
					}
				}
				gomega.Expect(hasCPU).To(gomega.BeTrue())
				gomega.Expect(nrt.TopologyPolicies[0]).ToNot(gomega.BeEmpty())
			}
		})
	})
})

func getNodeResourceTopology(tc *topologyclientset.Clientset, namespace, name string) *v1alpha1.NodeResourceTopology {
	var err error
	var nrt *v1alpha1.NodeResourceTopology
	gomega.EventuallyWithOffset(1, func() error {
		nrt, err = tc.TopologyV1alpha1().NodeResourceTopologies(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		return nil
	}, 1*time.Minute, 15*time.Second).ShouldNot(gomega.HaveOccurred())
	gomega.ExpectWithOffset(1, err).ToNot(gomega.HaveOccurred())
	return nrt
}

func waitPodsToBeRunningByRegex(pattern string) {
	gomega.EventuallyWithOffset(1, func() error {
		pods, err := getPodsByRegex(fmt.Sprintf("%s-*", pattern))
		if err != nil {
			return err
		}
		if len(pods) == 0 {
			return fmt.Errorf("no pods found for %q", pattern)
		}

		for _, pod := range pods {
			if pod.Status.Phase != corev1.PodRunning {
				return fmt.Errorf("pod %q is not in %v state", pod.Name, corev1.PodRunning)
			}
		}
		return nil
	}, 1*time.Minute, 15*time.Second).ShouldNot(gomega.HaveOccurred())
}

func getPodsByRegex(reg string) ([]*corev1.Pod, error) {
	cs, err := clientutil.New()
	if err != nil {
		return nil, err
	}

	podNameRgx, err := regexp.Compile(reg)
	if err != nil {
		return nil, err
	}

	podList := &corev1.PodList{}
	err = cs.List(context.TODO(), podList)
	if err != nil {
		return nil, err
	}

	ret := []*v1.Pod{}
	for _, pod := range podList.Items {
		if match := podNameRgx.FindString(pod.Name); len(match) != 0 {
			ret = append(ret, &pod)
		}
	}
	return ret, nil
}

func deploy() error {
	cmdline := []string{
		filepath.Join(binariesPath, "deployer"),
		"deploy",
	}
	fmt.Fprintf(ginkgo.GinkgoWriter, "running: %v\n", cmdline)

	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Stdout = ginkgo.GinkgoWriter
	cmd.Stderr = ginkgo.GinkgoWriter

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to deploy componentes before test started: %v", err)
	}
	return nil
}

func remove() error {
	cmdline := []string{
		filepath.Join(binariesPath, "deployer"),
		"remove",
	}
	fmt.Fprintf(ginkgo.GinkgoWriter, "running: %v\n", cmdline)

	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Stdout = ginkgo.GinkgoWriter
	cmd.Stderr = ginkgo.GinkgoWriter

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove components after test finished: %v", err)
	}
	return nil
}
