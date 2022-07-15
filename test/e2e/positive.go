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
	"sync"
	"time"

	appsv1 "k8s.io/api/apps/v1"
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
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform/detect"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/updaters"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests/nfd"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests/rte"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests/sched"
	"github.com/k8stopologyawareschedwg/deployer/pkg/tlog"
	"github.com/k8stopologyawareschedwg/deployer/pkg/validator"

	e2enodes "github.com/k8stopologyawareschedwg/deployer/test/e2e/utils/nodes"
	e2epods "github.com/k8stopologyawareschedwg/deployer/test/e2e/utils/pods"
	e2ewait "github.com/k8stopologyawareschedwg/deployer/test/e2e/utils/wait"
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
				"-P", "kubernetes:v1.24",
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

			do := detect.ClusterInfo{}
			if err := json.Unmarshal(out, &do); err != nil {
				ginkgo.Fail(fmt.Sprintf("Error unmarshalling output %q: %v", out, err))
			}
			gomega.Expect(do.Platform.AutoDetected).To(gomega.Equal(platform.Kubernetes))
			gomega.Expect(do.Platform.Discovered).To(gomega.Equal(platform.Kubernetes))

			minVer, err := platform.ParseVersion(validator.ExpectedMinKubeVersion)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(do.Version.Discovered.AtLeast(minVer)).To(gomega.BeTrue(), "cluster version mismatch - check validation")
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
			gomega.Expect(vo.Errors).To(gomega.BeEmpty(), "unexpected validation: %s", vo.String())
			gomega.Expect(vo.Success).To(gomega.BeTrue(), "unexpected validation: %s", vo.String())
		})
	})
})

var _ = ginkgo.Describe("[PositiveFlow] Deployer execution", func() {
	ginkgo.Context("with a running cluster without any components", func() {
		var updaterType string
		ginkgo.JustBeforeEach(func() {
			err := deploy(updaterType)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		})

		ginkgo.JustAfterEach(func() {
			err := remove(updaterType)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		})
		ginkgo.When("deployed with resource-topology-exporter as the updater", func() {
			ginkgo.BeforeEach(func() {
				updaterType = updaters.RTE
			})
			ginkgo.AfterEach(func() {
				updaterType = updaters.RTE
			})
			ginkgo.It("should perform overall deployment", func() {
				ginkgo.By("checking that resource-topology-exporter pod is running")

				ns, err := manifests.Namespace(manifests.ComponentResourceTopologyExporter)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				mf, err := rte.GetManifests(platform.Kubernetes, platform.Version("1.23"), ns.Name)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				mf = mf.Render(rte.RenderOptions{
					Namespace: ns.Name,
				})
				e2epods.WaitPodsToBeRunningByRegex(fmt.Sprintf("%s-*", mf.DaemonSet.Name))

				ginkgo.By("checking that topo-aware-scheduler pod is running")
				mfs, err := sched.GetManifests(platform.Kubernetes, ns.Name)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				// no need for options!
				mfs = mfs.Render(tlog.NewNullLogAdapter(), sched.RenderOptions{})
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
			})

			ginkgo.It("should verify a test pod scheduled with the topology aware scheduler goes running", func() {
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

		ginkgo.When("deployed with node-feature-discovery as the updater", func() {
			ginkgo.BeforeEach(func() {
				updaterType = updaters.NFD
			})
			ginkgo.AfterEach(func() {
				updaterType = updaters.NFD
			})
			ginkgo.It("should perform overall deployment", func() {
				ginkgo.By("checking that node-feature-discovery pods are running")

				ns, err := manifests.Namespace(manifests.ComponentNodeFeatureDiscovery)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				mf, err := nfd.GetManifests(platform.Kubernetes, ns.Name)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				mf = mf.Render(nfd.RenderOptions{
					Namespace: ns.Name,
				})
				e2epods.WaitPodsToBeRunningByRegex(fmt.Sprintf("%s-*", mf.DPMaster.Name))
				e2epods.WaitPodsToBeRunningByRegex(fmt.Sprintf("%s-*", mf.DSTopologyUpdater.Name))

				ginkgo.By("checking that topo-aware-scheduler pod is running")
				mfs, err := sched.GetManifests(platform.Kubernetes, ns.Name)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				// no need for options!
				mfs = mfs.Render(tlog.NewNullLogAdapter(), sched.RenderOptions{})
				e2epods.WaitPodsToBeRunningByRegex(fmt.Sprintf("%s-*", mfs.DPScheduler.Name))

				ginkgo.By("checking that noderesourcetopolgy has some information in it")
				tc, err := clientutil.NewTopologyClient()
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				workers, err := nodes.GetWorkers()
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				for _, node := range workers {
					ginkgo.By(fmt.Sprintf("checking node resource topology for %q", node.Name))

					// the name of the nrt object is the same as the worker node's name
					nrt := getNodeResourceTopology(tc, mf.DSTopologyUpdater.Namespace, node.Name)
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

			ginkgo.It("should verify a test pod scheduled with the topology aware scheduler goes running", func() {
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
})

var _ = ginkgo.Describe("[PositiveFlow] Deployer partial execution", func() {
	ginkgo.Context("with a running cluster without any components", func() {
		ginkgo.It("should perform the deployment of scheduler plugin + API and verify all pods are running", func() {
			binPath := filepath.Join(binariesPath, "deployer")

			err := runCmdline(
				[]string{binPath, "--debug", "deploy", "api", "--wait"},
				"failed to deploy partial components before test started",
			)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			err = runCmdline(
				[]string{binPath, "--debug", "deploy", "scheduler-plugin", "--wait"},
				"failed to deploy partial components before test started",
			)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			defer func() {
				err := runCmdline(
					[]string{binPath, "--debug", "remove", "scheduler-plugin", "--wait"},
					"failed to remove partial components after test finished",
				)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				err = runCmdline(
					[]string{binPath, "--debug", "remove", "api", "--wait"},
					"failed to remove partial components after test finished",
				)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
			}()

			ginkgo.By("checking that scheduler plugin is running")

			ns, err := manifests.Namespace(manifests.ComponentSchedulerPlugin)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			ginkgo.By("checking that topo-aware-scheduler pod is running")
			// TODO: autodetect the platform
			mfs, err := sched.GetManifests(platform.Kubernetes, ns.Name)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			// no need for options!
			mfs = mfs.Render(tlog.NewNullLogAdapter(), sched.RenderOptions{})

			cli, err := clientutil.New()
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			var wg sync.WaitGroup
			for _, dp := range []*appsv1.Deployment{
				mfs.DPScheduler,
				mfs.DPController,
			} {
				wg.Add(1)
				go func(dp *appsv1.Deployment) {
					defer ginkgo.GinkgoRecover()
					defer wg.Done()
					err = e2ewait.ForDeploymentComplete(cli, dp, 10*time.Second, 3*time.Minute)
					gomega.Expect(err).ToNot(gomega.HaveOccurred())
				}(dp)
			}
			wg.Wait()
		})
	})
})

func getNodeResourceTopology(tc *topologyclientset.Clientset, namespace, name string) *v1alpha1.NodeResourceTopology {
	var err error
	var nrt *v1alpha1.NodeResourceTopology
	fmt.Fprintf(ginkgo.GinkgoWriter, "looking for noderesourcetopology %q in namespace %q\n", name, namespace)
	gomega.EventuallyWithOffset(1, func() error {
		nrt, err = tc.TopologyV1alpha1().NodeResourceTopologies().Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		return nil
	}, 1*time.Minute, 15*time.Second).ShouldNot(gomega.HaveOccurred())
	gomega.ExpectWithOffset(1, err).ToNot(gomega.HaveOccurred())
	return nrt
}

func deploy(updaterType string) error {
	cmdline := []string{
		filepath.Join(binariesPath, "deployer"),
		"--debug",
		"deploy",
		"--rte-config-file", filepath.Join(deployerBaseDir, "hack", "rte.yaml"),
		"--wait",
	}
	if updaterType != "" {
		updaterArg := fmt.Sprintf("--updater-type=%s", updaterType)
		cmdline = append(cmdline, updaterArg)
	}
	// TODO: use error wrapping
	return runCmdline(cmdline, "failed to deploy components before test started")
}

func remove(updaterType string) error {
	cmdline := []string{
		filepath.Join(binariesPath, "deployer"),
		"--debug",
		"remove",
		"--wait",
	}
	if updaterType != "" {
		updaterArg := fmt.Sprintf("--updater-type=%s", updaterType)
		cmdline = append(cmdline, updaterArg)
	}
	// TODO: use error wrapping
	return runCmdline(cmdline, "failed to remove components after test finished")
}

func runCmdline(cmdline []string, errMsg string) error {
	fmt.Fprintf(ginkgo.GinkgoWriter, "running: %v\n", cmdline)

	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Stdout = ginkgo.GinkgoWriter
	cmd.Stderr = ginkgo.GinkgoWriter

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %v", errMsg, err)
	}
	return nil
}
