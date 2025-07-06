/*
 * Copyright 2025 Red Hat, Inc.
 *
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
 */

package e2e

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/k8stopologyawareschedwg/deployer/pkg/clientutil"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests/rte"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests/sched"
	"github.com/k8stopologyawareschedwg/deployer/pkg/options"

	e2epods "github.com/k8stopologyawareschedwg/deployer/test/e2e/utils/pods"
)

// This test suite verifies that the correct network policies are applied and enforced.
// Specifically, it checks the following:
// - All ingress and egress traffic is denied by default for the tas-scheduler and tas-topology-updater namespaces.
// - Egress traffic from the controller, RTE, and scheduler pods to the Kubernetes API server is allowed.
// - Ingress and egress traffic to/from other pods in the cluster is restricted.
//
// Full coverage for inter-pod communication is challenging, so we include basic tests to validate the expected behavior.

var _ = Describe("network policies are applied", Ordered, Label("feature:network_policies"), func() {
	ctx := context.Background()
	var cs client.Client
	var k8sClient *kubernetes.Clientset
	var controllerPod, schedulerPod, rteWorkerPod *corev1.Pod
	var err error

	BeforeAll(func() {
		cs, err = clientutil.New()
		Expect(err).ToNot(HaveOccurred())

		k8sClient, err = clientutil.NewK8s()
		Expect(err).ToNot(HaveOccurred())

		err = deployWithManifests()
		Expect(err).ToNot(HaveOccurred())
		By("checking that resource-topology-exporter pod is running")

		ns, err := manifests.Namespace(manifests.ComponentResourceTopologyExporter)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		enableCRIHooks := true
		mf, err := rte.NewWithOptions(options.Render{
			Platform:            platform.Kubernetes,
			PlatformVersion:     platform.Version("1.23"),
			Namespace:           ns.Name,
			EnableCRIHooks:      enableCRIHooks,
			CustomSELinuxPolicy: true,
		})
		Expect(err).ToNot(HaveOccurred())
		mf, err = mf.Render(options.UpdaterDaemon{
			Namespace: ns.Name,
		})
		Expect(err).ToNot(HaveOccurred())
		e2epods.WaitPodsToBeRunningByRegex(fmt.Sprintf("%s-*", mf.DaemonSet.Name))
		rteWorkerPods, err := e2epods.GetByRegex(cs, fmt.Sprintf("%s-*", mf.DaemonSet.Name))

		Expect(err).ToNot(HaveOccurred())
		Expect(rteWorkerPods).ToNot(BeEmpty())
		rteWorkerPod = rteWorkerPods[0]
		Expect(err).ToNot(HaveOccurred())

		By("checking that topo-aware-scheduler pod is running")
		mfs, err := sched.NewWithOptions(options.Render{
			Platform:  platform.Kubernetes,
			Namespace: ns.Name,
		})
		Expect(err).ToNot(HaveOccurred())
		mfs, err = mfs.Render(logr.Discard(), options.Scheduler{
			Replicas: int32(1),
		})
		Expect(err).ToNot(HaveOccurred())
		e2epods.WaitPodsToBeRunningByRegex(fmt.Sprintf("%s-*", mfs.DPScheduler.Name))
		schedulerPods, err := e2epods.GetByRegex(cs, fmt.Sprintf("%s-*", mfs.DPScheduler.Name))
		Expect(err).ToNot(HaveOccurred())
		Expect(schedulerPods).ToNot(BeEmpty())
		schedulerPod = schedulerPods[0]

		By("checking that topo-aware-controller is running")
		e2epods.WaitPodsToBeRunningByRegex(fmt.Sprintf("%s-*", mfs.DPController.Name))
		controllerPods, err := e2epods.GetByRegex(cs, fmt.Sprintf("%s-*", mfs.DPController.Name))
		Expect(err).ToNot(HaveOccurred())
		Expect(controllerPods).ToNot(BeEmpty())
		controllerPod = controllerPods[0]

		By("checking that noderesourcetopolgy has some information in it")
		expectNodeResourceTopologyData()
	})
	AfterAll(func() {
		err := remove("")
		Expect(err).ToNot(HaveOccurred())
	})
	type trafficCase struct {
		FromPod     func() *corev1.Pod
		ToHost      func() string
		ToPort      string
		ShouldAllow bool
		Description string
	}

	DescribeTable("traffic behavior",
		func(tc trafficCase) {
			Expect(tc.FromPod).ToNot(BeNil(), "source pod should not be nil")
			klog.InfoS("Running traffic test", "description", tc.Description)
			reachable := trafficTest(k8sClient, ctx, tc.FromPod(), tc.ToHost(), tc.ToPort)
			klog.InfoS("reachable", "reachable", reachable)

			if tc.ShouldAllow {
				Expect(reachable).To(BeTrue(), tc.Description)
			} else {
				Expect(reachable).To(BeFalse(), tc.Description)
			}
		},
		// Testing controller, rte, and scheduler egress traffic to API server
		Entry("controller -> API server", trafficCase{
			FromPod:     func() *corev1.Pod { return controllerPod },
			ToHost:      func() string { return "$KUBERNETES_SERVICE_HOST" },
			ToPort:      "$KUBERNETES_SERVICE_PORT",
			ShouldAllow: true,
			Description: "controller should access the API server",
		}),
		Entry("scheduler -> API server", trafficCase{
			FromPod:     func() *corev1.Pod { return schedulerPod },
			ToHost:      func() string { return "$KUBERNETES_SERVICE_HOST" },
			ToPort:      "$KUBERNETES_SERVICE_PORT",
			ShouldAllow: true,
			Description: "scheduler should access the API server",
		}),
		Entry("rte worker -> API server", trafficCase{
			FromPod:     func() *corev1.Pod { return rteWorkerPod },
			ToHost:      func() string { return "$KUBERNETES_SERVICE_HOST" },
			ToPort:      "$KUBERNETES_SERVICE_PORT",
			ShouldAllow: true,
			Description: "rte worker should access the API server",
		}),
		// Testing traffic restrictions between pods
		Entry("controller -> scheduler", trafficCase{
			FromPod:     func() *corev1.Pod { return controllerPod },
			ToHost:      func() string { return schedulerPod.Status.PodIP },
			ToPort:      "10259", // liveness probe
			ShouldAllow: false,
			Description: "controller should NOT access scheduler's liveness probe",
		}),
	)
})

// trafficTest returns true if the sourcePod can connect to the given destination IP and port using Ncat.
func trafficTest(cli *kubernetes.Clientset, ctx context.Context, sourcePod *corev1.Pod, destinationIP, destinationPort string) bool {
	GinkgoHelper()

	endpoint := net.JoinHostPort(destinationIP, destinationPort)
	key := client.ObjectKeyFromObject(sourcePod)

	By(fmt.Sprintf("testing network connectivity from pod %q to %s", key.String(), endpoint))

	cmd := []string{
		"sh", "-c",
		// nc compatible format is <host> <port> (without colon)
		fmt.Sprintf(`nc -w 5 -z %s %s && echo "OK" || echo "FAIL"`, destinationIP, destinationPort),
	}

	out, _ := e2epods.ExecCommand(cli, ctx, sourcePod, "", cmd)
	output := string(out)

	klog.InfoS("traffic tcp connection", "pod", key.String(), "destination", endpoint, "output", output)

	return strings.Contains(output, "OK")
}
