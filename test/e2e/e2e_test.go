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
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha2"
	nrtattrv1alpha2 "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha2/helper/attribute"
	topologyclientset "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/generated/clientset/versioned"

	"github.com/k8stopologyawareschedwg/podfingerprint"

	"github.com/k8stopologyawareschedwg/deployer/pkg/clientutil"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/stringify"
	"github.com/k8stopologyawareschedwg/deployer/pkg/validator"
)

var (
	deployerBaseDir string
	binariesPath    string
)

func TestE2E(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Deployer Suite")
}

var _ = ginkgo.BeforeSuite(func() {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		ginkgo.Fail("Cannot retrieve tests directory")
	}
	basedir := filepath.Dir(file)
	deployerBaseDir = filepath.Clean(filepath.Join(basedir, "..", ".."))
	binariesPath = filepath.Clean(filepath.Join(deployerBaseDir, "_out"))
	fmt.Fprintf(ginkgo.GinkgoWriter, "using binaries at %q\n", binariesPath)
})

type validationOutput struct {
	Success bool                         `json:"success"`
	Errors  []validator.ValidationResult `json:"errors,omitempty"`
}

func (vo validationOutput) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "validation: success=%t\n", vo.Success)
	for _, vErr := range vo.Errors {
		fmt.Fprintf(&sb, "validation: error: %s\n", vErr.String())
	}
	return sb.String()
}

type imageOutput struct {
	TopologyUpdater     string `json:"topology_updater"`
	SchedulerPlugin     string `json:"scheduler_plugin"`
	SchedulerController string `json:"scheduler_controller"`
}

func checkHasCPU(nrt *v1alpha2.NodeResourceTopology) error {
	// we check CPUs because that's the only resource we know it will always be available
	ok := false
	for _, zone := range nrt.Zones {
		for _, resource := range zone.Resources {
			if resource.Name == string(corev1.ResourceCPU) && resource.Capacity.Size() >= 1 {
				ok = true
			}
		}
	}
	if !ok {
		return fmt.Errorf("missing CPUs in %q", nrt.Name)
	}
	return nil
}

func checkHasPFP(nrt *v1alpha2.NodeResourceTopology) error {
	_, hasAttr := nrtattrv1alpha2.Get(nrt.Attributes, podfingerprint.Attribute)
	if !hasAttr {
		return fmt.Errorf("PFP attribute missing in %q", nrt.Name)
	}
	// TODO: check annotation _only for RTE_?
	return nil
}

func checkLacksPFP(nrt *v1alpha2.NodeResourceTopology) error {
	attr, hasAttr := nrtattrv1alpha2.Get(nrt.Attributes, podfingerprint.Attribute)
	if hasAttr {
		return fmt.Errorf("PFP attribute found: %v in %q", attr.Value, nrt.Name)
	}
	valAnn, hasAnn := nrt.Annotations[podfingerprint.Annotation]
	if hasAnn {
		return fmt.Errorf("PFP annotation found: %v in %q", valAnn, nrt.Name)
	}
	return nil
}

func getNodeResourceTopology(tc topologyclientset.Interface, name string, filterFunc func(nrt *v1alpha2.NodeResourceTopology) error) *v1alpha2.NodeResourceTopology {
	var err error
	var nrt *v1alpha2.NodeResourceTopology
	fmt.Fprintf(ginkgo.GinkgoWriter, "looking for noderesourcetopology %q\n", name)
	err = wait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
		nrt, err = tc.TopologyV1alpha2().NodeResourceTopologies().Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		err = filterFunc(nrt)
		if err != nil {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		dumpNRT(tc)
	}
	return nrt
}

func ensureNodeResourceTopology(tc topologyclientset.Interface, name string, filterFunc func(nrt *v1alpha2.NodeResourceTopology) error) error {
	fmt.Fprintf(ginkgo.GinkgoWriter, "ensuring predicate for noderesourcetopology %q\n", name)
	var err error
	var nrt *v1alpha2.NodeResourceTopology
	for attempt := 0; attempt <= 12; attempt++ {
		nrt, err = tc.TopologyV1alpha2().NodeResourceTopologies().Get(context.TODO(), name, metav1.GetOptions{})
		if err != nil {
			break
		}
		err = filterFunc(nrt)
		if err != nil {
			break
		}

		time.Sleep(5 * time.Second)
	}
	if err != nil {
		dumpNRT(tc)
	}
	return err
}

func dumpNRT(tc topologyclientset.Interface) {
	nrts, err := tc.TopologyV1alpha2().NodeResourceTopologies().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Fprintf(ginkgo.GinkgoWriter, "cannot dump NRTs in the cluster: %v\n", err)
		return
	}
	fmt.Fprintf(ginkgo.GinkgoWriter, "%s\n", stringify.NodeResourceTopologyList(nrts.Items, "cluster NRTs"))
}

func deployWithManifests() error {
	cmdline := []string{
		"kubectl",
		"create",
		"-f",
		filepath.Join(binariesPath, "deployer-manifests-allinone.yaml"),
	}
	// TODO: use error wrapping
	err := runCmdline(cmdline, "failed to deploy components before test started")
	if err != nil {
		dumpSchedulerPods()
	}
	return err
}

func deploy(updaterType string, pfpEnable bool) error {
	cmdline := []string{
		filepath.Join(binariesPath, "deployer"),
		"deploy",
		"--rte-config-file=" + filepath.Join(deployerBaseDir, "hack", "rte.yaml"),
		"--updater-pfp-enable=" + strconv.FormatBool(pfpEnable),
		"--sched-ctrlplane-affinity=false",
		"--wait",
	}
	if updaterType != "" {
		updaterArg := fmt.Sprintf("--updater-type=%s", updaterType)
		cmdline = append(cmdline, updaterArg)
	}
	// TODO: use error wrapping
	err := runCmdline(cmdline, "failed to deploy components before test started")
	if err != nil {
		dumpSchedulerPods()
	}
	return err
}

func remove(updaterType string) error {
	cmdline := []string{
		filepath.Join(binariesPath, "deployer"),
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
		return fmt.Errorf("%s: %w", errMsg, err)
	}
	return nil
}

func NullEnv() *deployer.Environment {
	cli, err := clientutil.New()
	gomega.ExpectWithOffset(1, err).ToNot(gomega.HaveOccurred())
	env := deployer.Environment{
		Ctx: context.TODO(),
		Cli: cli,
		Log: logr.Discard(),
	}
	return &env
}
