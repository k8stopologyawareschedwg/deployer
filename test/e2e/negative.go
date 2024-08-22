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
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/go-logr/logr"
	"github.com/k8stopologyawareschedwg/deployer/pkg/clientutil"
	"github.com/k8stopologyawareschedwg/deployer/pkg/clientutil/nodes"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/updaters"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests/nfd"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests/rte"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests/sched"
	"github.com/k8stopologyawareschedwg/deployer/pkg/options"
	e2epods "github.com/k8stopologyawareschedwg/deployer/test/e2e/utils/pods"
)

var _ = ginkgo.Describe("[NegativeFlow] Deployer validation", ginkgo.Label("negative"), func() {
	ginkgo.Context("with cluster with default settings", func() {
		ginkgo.It("it should fail the validation", func() {
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
			gomega.Expect(vo.Success).To(gomega.BeFalse())
			gomega.Expect(vo.Errors).ToNot(gomega.BeEmpty())
		})
	})
})

var _ = ginkgo.Describe("[NegativeFlow] Deployer option validation", ginkgo.Label("negative"), func() {
	ginkgo.It("It should fail with invalid --updater-type", func() {
		updaterType := "LOCAL"
		err := deploy(updaterType, true)
		gomega.Expect(err).To(gomega.HaveOccurred(), "deployed succesfully with unknown updater type %s", updaterType)
	})
})

var _ = ginkgo.Describe("[NegativeFlow] Deployer execution with PFP disabled", ginkgo.Label("negative"), func() {
	ginkgo.Context("with a running cluster without any components", func() {
		var updaterType string
		ginkgo.JustBeforeEach(func() {
			err := deploy(updaterType, false)
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

				enableCRIHooks := true
				mf, err := rte.GetManifests(platform.Kubernetes, platform.Version("1.23"), ns.Name, enableCRIHooks)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				mf, err = mf.Render(options.UpdaterDaemon{
					Namespace: ns.Name,
				})
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				e2epods.WaitPodsToBeRunningByRegex(fmt.Sprintf("%s-*", mf.DaemonSet.Name))

				ginkgo.By("checking that topo-aware-scheduler pod is running")
				mfs, err := sched.GetManifests(platform.Kubernetes, ns.Name)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				mfs, err = mfs.Render(logr.Discard(), options.Scheduler{
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
					err = ensureNodeResourceTopology(tc, node.Name, checkLacksPFP)
					gomega.Expect(err).ToNot(gomega.HaveOccurred())
				}
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
					mf, err = mf.Render(options.UpdaterDaemon{
						Namespace: ns.Name,
					})
					gomega.Expect(err).ToNot(gomega.HaveOccurred())
					e2epods.WaitPodsToBeRunningByRegex(fmt.Sprintf("%s-*", mf.DSTopologyUpdater.Name))

					ginkgo.By("checking that topo-aware-scheduler pod is running")
					mfs, err := sched.GetManifests(platform.Kubernetes, ns.Name)
					gomega.Expect(err).ToNot(gomega.HaveOccurred())
					mfs, err = mfs.Render(logr.Discard(), options.Scheduler{
						Replicas: int32(1),
					})
					gomega.Expect(err).ToNot(gomega.HaveOccurred())
					e2epods.WaitPodsToBeRunningByRegex(fmt.Sprintf("%s-*", mfs.DPScheduler.Name))

					ginkgo.By("checking that noderesourcetopolgy has some information in it")
					expectNodeResourceTopologyData()
				})
			})
		})
	})
})
