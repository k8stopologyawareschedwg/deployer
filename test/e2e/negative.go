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

	"github.com/k8stopologyawareschedwg/deployer/pkg/clientutil"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform/detect"
	rtedeploy "github.com/k8stopologyawareschedwg/deployer/pkg/deployer/rte"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests/api"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests/rte"
	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

var _ = ginkgo.Describe("[NegativeFlow] Deployer validation", func() {
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

		ginkgo.It("it should not have any manifest", func() {
			dp, err := detect.Detect()
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			cli, err := clientutil.New()
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			apiMf, err := api.GetManifests(dp)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			exApiMf := apiMf.FromClient(context.TODO(), cli)
			gomega.Expect(exApiMf.Existing.Crd).To(gomega.BeNil())
			gomega.Expect(checkError(exApiMf.CrdError)).ToNot(gomega.HaveOccurred(), "unexpected err: %v", err)

			rteMf, err := rte.GetManifests(dp)
			_, ns, err := rtedeploy.SetupNamespace(dp)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			rteMf = rteMf.Update(rte.UpdateOptions{Namespace: ns})
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			exRteMf := rteMf.FromClient(context.TODO(), cli)

			gomega.Expect(exRteMf.Existing.DaemonSet).To(gomega.BeNil())
			gomega.Expect(exRteMf.Existing.ServiceAccount).To(gomega.BeNil())
			gomega.Expect(exRteMf.Existing.Role).To(gomega.BeNil())
			gomega.Expect(exRteMf.Existing.RoleBinding).To(gomega.BeNil())
			// ConfigMap can be missing, and it's ok

			gomega.Expect(checkError(exRteMf.DaemonSetError)).ToNot(gomega.HaveOccurred(), "unexpected err: %v", err)
			gomega.Expect(checkError(exRteMf.ServiceAccountError)).ToNot(gomega.HaveOccurred(), "unexpected err: %v", err)
			gomega.Expect(checkError(exRteMf.RoleError)).ToNot(gomega.HaveOccurred(), "unexpected err: %v", err)
			gomega.Expect(checkError(exRteMf.RoleBindingError)).ToNot(gomega.HaveOccurred(), "unexpected err: %v", err)
			// ConfigMap can be missing, and it's ok

			// TODO: add sched
		})
	})
})

func checkError(err error) error {
	if err == nil {
		return nil
	}
	if apierrors.IsNotFound(err) {
		fmt.Fprintf(ginkgo.GinkgoWriter, "isNotFound!\n")
		return nil
	}
	// TODO: better to use wrap
	return fmt.Errorf("error is not nil nor IsNotFound: %v", err)
}
