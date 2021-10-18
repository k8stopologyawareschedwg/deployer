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
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha1"
	topologyclientset "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/generated/clientset/versioned"

	"github.com/hashicorp/go-version"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"github.com/k8stopologyawareschedwg/deployer/pkg/clientutil"
	"github.com/k8stopologyawareschedwg/deployer/pkg/clientutil/nodes"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests/rte"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests/sched"
	"github.com/k8stopologyawareschedwg/deployer/pkg/tlog"

	e2enodes "github.com/k8stopologyawareschedwg/deployer/test/e2e/utils/nodes"
	e2epods "github.com/k8stopologyawareschedwg/deployer/test/e2e/utils/pods"
)

var _ = ginkgo.Describe("[PositiveFlow] Deployer version", func() {
	ginkgo.Context("with the tool available", func() {
		ginkgo.It("it should show the correct version", func() {
			cmdline := []string{
				filepath.Join(binariesPath, "deployer"),
				"version",
			}
			fmt.Fprintf(ginkgo.GinkgoWriter, "running: %v\n", cmdline)

			cmd := exec.Command(cmdline[0], cmdline[1:]...)
			cmd.Stderr = ginkgo.GinkgoWriter

			out, err := cmd.Output()
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			text := strings.TrimSpace(string(out))
			fmt.Fprintf(ginkgo.GinkgoWriter, "reported version: %q\n", text)
			gomega.Expect(text).ToNot(gomega.BeEmpty())
			_, err = version.NewVersion(text)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		})
	})
})

var _ = ginkgo.Describe("[PositiveFlow] Deployer images", func() {
	ginkgo.Context("with the tool available", func() {
		ginkgo.It("it should emit the images being used", func() {
			cmdline := []string{
				filepath.Join(binariesPath, "deployer"),
				"images",
				"--json",
			}
			fmt.Fprintf(ginkgo.GinkgoWriter, "running: %v\n", cmdline)

			cmd := exec.Command(cmdline[0], cmdline[1:]...)
			cmd.Stderr = ginkgo.GinkgoWriter

			out, err := cmd.Output()
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			imo := imageOutput{}
			if err := json.Unmarshal(out, &imo); err != nil {
				ginkgo.Fail(fmt.Sprintf("Error unmarshalling output %q: %v", out, err))
			}

			gomega.Expect(imo.TopologyUpdater).ToNot(gomega.BeNil())
			gomega.Expect(imo.SchedulerPlugin).ToNot(gomega.BeNil())
			gomega.Expect(imo.SchedulerController).ToNot(gomega.BeNil())
		})
	})
})

var _ = ginkgo.Describe("[PositiveFlow] Deployer render", func() {
	ginkgo.Context("with cluster image overrides", func() {
		ginkgo.It("it should reflect the overrides in the output", func() {
			cmdline := []string{
				filepath.Join(binariesPath, "deployer"),
				"-P", "kubernetes",
				"render",
			}
			fmt.Fprintf(ginkgo.GinkgoWriter, "running: %v\n", cmdline)

			testSchedPlugImage := "quay.io/sched/sched:test000"
			testResTopoExImage := "quay.io/rte/rte:test000"

			cmd := exec.Command(cmdline[0], cmdline[1:]...)
			cmd.Stderr = ginkgo.GinkgoWriter
			cmd.Env = append(cmd.Env, fmt.Sprintf("TAS_SCHEDULER_PLUGIN_IMAGE=%s", testSchedPlugImage))
			cmd.Env = append(cmd.Env, fmt.Sprintf("TAS_RESOURCE_EXPORTER_IMAGE=%s", testResTopoExImage))

			out, err := cmd.Output()
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			text := string(out)
			gomega.Expect(strings.Contains(text, testSchedPlugImage)).To(gomega.BeTrue())
			gomega.Expect(strings.Contains(text, testResTopoExImage)).To(gomega.BeTrue())
		})
	})
})

var _ = ginkgo.Describe("[PositiveFlow] Deployer detection", func() {
	ginkgo.Context("with cluster with the expected settings", func() {
		ginkgo.It("it should detect a kubernetes cluster as such", func() {
			cmdline := []string{
				filepath.Join(binariesPath, "deployer"),
				"detect",
				"--json",
			}
			fmt.Fprintf(ginkgo.GinkgoWriter, "running: %v\n", cmdline)

			cmd := exec.Command(cmdline[0], cmdline[1:]...)
			cmd.Stderr = ginkgo.GinkgoWriter

			out, err := cmd.Output()
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			do := detectionOutput{}
			if err := json.Unmarshal(out, &do); err != nil {
				ginkgo.Fail(fmt.Sprintf("Error unmarshalling output %q: %v", out, err))
			}
			gomega.Expect(do.AutoDetected).To(gomega.Equal(platform.Kubernetes))
			gomega.Expect(do.Discovered).To(gomega.Equal(platform.Kubernetes))
		})
	})
})

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

			ns, err := manifests.Namespace(manifests.ComponentResourceTopologyExporter)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			mf, err := rte.GetManifests(platform.Kubernetes)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			mf = mf.Update(rte.UpdateOptions{
				Namespace: ns.Name,
			})
			e2epods.WaitPodsToBeRunningByRegex(fmt.Sprintf("%s-*", mf.DaemonSet.Name))

			ginkgo.By("checking that topo-aware-scheduler pod is running")
			mfs, err := sched.GetManifests(platform.Kubernetes)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			// no need for options!
			mfs = mfs.Update(tlog.NewNullLogAdapter(), sched.UpdateOptions{})
			e2epods.WaitPodsToBeRunningByRegex(fmt.Sprintf("%s-*", mfs.DPScheduler.Name))

			ginkgo.By("checking that noderesourcetopolgy has some information in it")
			tc, err := clientutil.NewTopologyClient()
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			workers, err := nodes.GetWorkers()
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			for _, node := range workers {
				ginkgo.By(fmt.Sprintf("checking node resource topology for %q", node.Name))

				// the name of the nrt object is the same as the worker node's name
				nrt := getNodeResourceTopology(tc, mf.DaemonSet.Namespace, node.Name)
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

			ginkgo.By("checking the cluster resource availability")
			cli, err := clientutil.NewK8s()
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			workerNodes, err := e2enodes.GetWorkerNodes(cli)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			if len(workerNodes) < 1 {
				// how come did the validation pass?
				ginkgo.Fail("no worker nodes found in the cluster")
			}

			// min 1 reserved + min 1 allocatable = 2
			nodes, err := e2enodes.FilterNodesWithEnoughCores(workerNodes, "2")
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			if len(nodes) < 1 {
				// TODO: it is unusual to skip so late, maybe split this spec in 2?
				ginkgo.Skip("skipping the pod check - not enough resources")
			}

			testNs := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "tas-test-",
				},
			}
			ginkgo.By("creating a test namespace")
			testNs, err = cli.CoreV1().Namespaces().Create(context.TODO(), testNs, metav1.CreateOptions{})
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			defer func() {
				cli.CoreV1().Namespaces().Delete(context.TODO(), testNs.Name, metav1.DeleteOptions{})
			}()

			// TODO autodetect the scheduler name
			testPod := e2epods.GuaranteedSleeperPod(testNs.Name, "topology-aware-scheduler")
			ginkgo.By("creating a guaranteed sleeper pod using the topology aware scheduler")
			testPod, err = cli.CoreV1().Pods(testPod.Namespace).Create(context.TODO(), testPod, metav1.CreateOptions{})
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			ginkgo.By("checking the pod goes running")
			e2epods.WaitForPodToBeRunning(cli, testPod.Namespace, testPod.Name)
		})
	})
})

func getNodeResourceTopology(tc *topologyclientset.Clientset, namespace, name string) *v1alpha1.NodeResourceTopology {
	var err error
	var nrt *v1alpha1.NodeResourceTopology
	fmt.Fprintf(ginkgo.GinkgoWriter, "looking for noderesourcetopology %q in namespace %q\n", name, namespace)
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

func deploy() error {
	cmdline := []string{
		filepath.Join(binariesPath, "deployer"),
		"--debug",
		"deploy",
		"--rte-config-file", filepath.Join(deployerBaseDir, "hack", "rte.yaml"),
		"--wait",
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
		"--debug",
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
