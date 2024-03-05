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
	"sync"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8swait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"

	"github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha2"
	nrtattrv1alpha2 "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha2/helper/attribute"
	topologyclientset "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/generated/clientset/versioned"

	"github.com/k8stopologyawareschedwg/podfingerprint"

	"github.com/k8stopologyawareschedwg/deployer/pkg/clientutil"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/wait"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests/sched"
	"github.com/k8stopologyawareschedwg/deployer/pkg/options"
	"github.com/k8stopologyawareschedwg/deployer/pkg/stringify"
	"github.com/k8stopologyawareschedwg/deployer/pkg/validator"
	e2epods "github.com/k8stopologyawareschedwg/deployer/test/e2e/utils/pods"
)

var (
	deployerBaseDir string
	binariesPath    string
)

func TestE2E(t *testing.T) {
	ctrllog.SetLogger(klogr.NewWithOptions(klogr.WithFormat(klogr.FormatKlog)))
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
	err = k8swait.PollImmediate(5*time.Second, 1*time.Minute, func() (bool, error) {
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
		"--updater-verbose=5",
		"--sched-ctrlplane-affinity=false",
		"--sched-verbose=5",
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

func dumpSchedulerPods() {
	ns, err := manifests.Namespace(manifests.ComponentSchedulerPlugin)
	gomega.ExpectWithOffset(1, err).ToNot(gomega.HaveOccurred())

	// TODO: autodetect the platform
	mfs, err := sched.GetManifests(platform.Kubernetes, ns.Name)
	gomega.ExpectWithOffset(1, err).ToNot(gomega.HaveOccurred())
	mfs, err = mfs.Render(logr.Discard(), options.Scheduler{
		Replicas: int32(1),
	})
	gomega.ExpectWithOffset(1, err).ToNot(gomega.HaveOccurred())

	cli, err := clientutil.New()
	gomega.ExpectWithOffset(1, err).ToNot(gomega.HaveOccurred())

	k8sCli, err := clientutil.NewK8s()
	gomega.ExpectWithOffset(1, err).ToNot(gomega.HaveOccurred())

	ctx := context.Background()

	pods, err := e2epods.GetByDeployment(cli, ctx, *mfs.DPScheduler)
	gomega.ExpectWithOffset(1, err).ToNot(gomega.HaveOccurred())

	klog.Warning(">>> scheduler pod status begin:\n")
	for idx := range pods {
		pod := &pods[idx]

		// TODO
		pod.ManagedFields = nil
		// TODO

		data, err := yaml.Marshal(pod)
		gomega.ExpectWithOffset(1, err).ToNot(gomega.HaveOccurred())

		klog.Warningf("%s\n---\n", string(data))

		e2epods.LogEventsForPod(k8sCli, ctx, pod.Namespace, pod.Name)
		klog.Warningf("---\n")
	}
	klog.Warning(">>> scheduler pod status end\n")
}

func expectSchedulerRunning() {
	ns, err := manifests.Namespace(manifests.ComponentSchedulerPlugin)
	gomega.ExpectWithOffset(1, err).ToNot(gomega.HaveOccurred())

	cli, err := clientutil.New()
	gomega.ExpectWithOffset(1, err).ToNot(gomega.HaveOccurred())

	ctx := context.Background()

	ginkgo.By("checking that scheduler plugin is configured")

	confMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns.Name,
			Name:      "scheduler-config", // TODO: duplicate from YAML
		},
	}
	err = cli.Get(ctx, client.ObjectKeyFromObject(&confMap), &confMap)
	gomega.ExpectWithOffset(1, err).ToNot(gomega.HaveOccurred())
	gomega.ExpectWithOffset(1, confMap.Data).ToNot(gomega.BeNil(), "empty config map for scheduler config")

	data, ok := confMap.Data[manifests.SchedulerConfigFileName]
	gomega.ExpectWithOffset(1, ok).To(gomega.BeTrue(), "empty config data for %q", manifests.SchedulerConfigFileName)

	allParams, err := manifests.DecodeSchedulerProfilesFromData([]byte(data))
	gomega.ExpectWithOffset(1, len(allParams)).To(gomega.Equal(1), "unexpected params: %#v", allParams)

	params := allParams[0] // TODO: smarter find
	gomega.ExpectWithOffset(1, err).ToNot(gomega.HaveOccurred())
	gomega.ExpectWithOffset(1, params.Cache).ToNot(gomega.BeNil(), "no data for scheduler cache config")
	gomega.ExpectWithOffset(1, params.Cache.ResyncPeriodSeconds).ToNot(gomega.BeNil(), "no data for scheduler cache resync period")

	ginkgo.By("checking that scheduler plugin is running")

	ginkgo.By("checking that topo-aware-scheduler pod is running")
	// TODO: autodetect the platform
	mfs, err := sched.GetManifests(platform.Kubernetes, ns.Name)
	gomega.ExpectWithOffset(1, err).ToNot(gomega.HaveOccurred())
	mfs, err = mfs.Render(logr.Discard(), options.Scheduler{
		Replicas: int32(1),
	})
	gomega.ExpectWithOffset(1, err).ToNot(gomega.HaveOccurred())

	var wg sync.WaitGroup
	for _, dp := range []*appsv1.Deployment{
		mfs.DPScheduler,
		mfs.DPController,
	} {
		wg.Add(1)
		go func(dp *appsv1.Deployment) {
			defer ginkgo.GinkgoRecover()
			defer wg.Done()
			_, err = wait.With(cli, logr.Discard()).Interval(10*time.Second).Timeout(3*time.Minute).ForDeploymentComplete(ctx, dp)
			gomega.ExpectWithOffset(1, err).ToNot(gomega.HaveOccurred())
		}(dp)
	}
	wg.Wait()
}
