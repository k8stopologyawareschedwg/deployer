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
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hashicorp/go-version"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("[PositiveFlow][Local] Deployer version", func() {
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

var _ = ginkgo.Describe("[PositiveFlow][Local] Deployer images", func() {
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

		ginkgo.It("it should enable to override the builtin images (scheduler)", func() {
			testImageSpec := "quay.io/foobar/sched:dev"
			gomega.Expect(os.Setenv("TAS_SCHEDULER_PLUGIN_IMAGE", testImageSpec)).To(gomega.Succeed())
			defer func() {
				ginkgo.GinkgoHelper()
				gomega.Expect(os.Unsetenv("TAS_SCHEDULER_PLUGIN_IMAGE")).To(gomega.Succeed())
			}()

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

			gomega.Expect(imo.SchedulerPlugin).To(gomega.Equal(testImageSpec))
		})

		ginkgo.It("it should enable to override the builtin images (controller)", func() {
			testImageSpec := "quay.io/foobar/ctrl:dev"
			gomega.Expect(os.Setenv("TAS_SCHEDULER_PLUGIN_CONTROLLER_IMAGE", testImageSpec)).To(gomega.Succeed())
			defer func() {
				ginkgo.GinkgoHelper()
				gomega.Expect(os.Unsetenv("TAS_SCHEDULER_PLUGIN_CONTROLLER_IMAGE")).To(gomega.Succeed())
			}()

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

			gomega.Expect(imo.SchedulerController).To(gomega.Equal(testImageSpec))
		})
	})
})

var _ = ginkgo.Describe("[PositiveFlow][Local] Deployer render", func() {
	ginkgo.Context("with default settings", func() {
		ginkgo.Context("with focus on topology-updater", func() {
			ginkgo.DescribeTable("pods fingerprinting support",
				func(updaterType string, expected bool) {
					cmdline := []string{
						filepath.Join(binariesPath, "deployer"),
						"-P", "kubernetes:v1.26",
						"--updater-type=" + updaterType,
						"--updater-pfp-enable=" + strconv.FormatBool(expected),
						"render",
					}
					fmt.Fprintf(ginkgo.GinkgoWriter, "running: %v\n", cmdline)

					cmd := exec.Command(cmdline[0], cmdline[1:]...)
					cmd.Stderr = ginkgo.GinkgoWriter
					out, err := cmd.Output()
					gomega.Expect(err).ToNot(gomega.HaveOccurred())

					text := string(out)
					// TODO: pretty crude. We should do something smarter
					haveFlag := strings.Contains(text, fmt.Sprintf("--pods-fingerprint=%v", strconv.FormatBool(expected)))
					desc := fmt.Sprintf("pods fingerprinting setting found=%v", haveFlag)
					gomega.Expect(haveFlag).To(gomega.BeTrue(), desc)
				},
				ginkgo.Entry("RTE pfp on", "RTE", true),
				ginkgo.Entry("NFD pfp on", "NFD", true),
				ginkgo.Entry("RTE pfp off", "RTE", false),
				ginkgo.Entry("NFD pfp off", "NFD", false),
			)
		})
	})

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
