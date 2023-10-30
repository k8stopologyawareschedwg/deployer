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

package e2e

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha2"

	"github.com/k8stopologyawareschedwg/deployer/pkg/clientutil"
	"github.com/k8stopologyawareschedwg/deployer/pkg/clientutil/nodes"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests/rte"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests/sched"

	e2enodes "github.com/k8stopologyawareschedwg/deployer/test/e2e/utils/nodes"
	e2epods "github.com/k8stopologyawareschedwg/deployer/test/e2e/utils/pods"
)

var _ = ginkgo.Describe("[ManifestFlow] Deployer rendering", func() {
	ginkgo.Context("with a running cluster without any components", func() {
		ginkgo.BeforeEach(func() {
			err := deployWithManifests()
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		})

		ginkgo.AfterEach(func() {
			err := remove("")
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
		})

		ginkgo.When("deployed using manifests", func() {
			ginkgo.It("should perform overall deployment", func() {
				ginkgo.By("checking that resource-topology-exporter pod is running")

				ns, err := manifests.Namespace(manifests.ComponentResourceTopologyExporter)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				enableCRIHooks := true
				mf, err := rte.GetManifests(platform.Kubernetes, platform.Version("1.23"), ns.Name, enableCRIHooks)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				mf, err = mf.Render(rte.RenderOptions{
					Namespace: ns.Name,
				})
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				e2epods.WaitPodsToBeRunningByRegex(fmt.Sprintf("%s-*", mf.DaemonSet.Name))

				ginkgo.By("checking that topo-aware-scheduler pod is running")
				mfs, err := sched.GetManifests(platform.Kubernetes, ns.Name)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				mfs, err = mfs.Render(logr.Discard(), sched.RenderOptions{
					Replicas: int32(1),
				})
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				e2epods.WaitPodsToBeRunningByRegex(fmt.Sprintf("%s-*", mfs.DPScheduler.Name))

				ginkgo.By("checking that noderesourcetopolgy has some information in it")
				tc, err := clientutil.NewTopologyClient()
				gomega.Expect(err).ToNot(gomega.HaveOccurred())

				workers, err := nodes.GetWorkers(NullEnv())
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				for _, node := range workers {
					ginkgo.By(fmt.Sprintf("checking node resource topology for %q", node.Name))

					// the name of the nrt object is the same as the worker node's name
					_ = getNodeResourceTopology(tc, node.Name, func(nrt *v1alpha2.NodeResourceTopology) error {
						if err := checkHasCPU(nrt); err != nil {
							return err
						}
						if err := checkHasPFP(nrt); err != nil {
							return err
						}
						return nil
					})
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
