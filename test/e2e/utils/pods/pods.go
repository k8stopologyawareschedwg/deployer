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

package pods

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8swait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/k8stopologyawareschedwg/deployer/pkg/clientutil"
)

const (
	CentosImage = "quay.io/centos/centos:8"
)

func GuaranteedSleeperPod(namespace, schedulerName string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sleeper-gu-pod",
			Namespace: namespace,
		},
		Spec: corev1.PodSpec{
			SchedulerName: schedulerName,
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:  "sleeper-gu-cnt",
					Image: CentosImage,
					// 1 hour (or >= 1h in general) is "forever" for our purposes
					Command: []string{"/bin/sleep", "1h"},
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceName(corev1.ResourceCPU): resource.MustParse("1"),
							// any random reasonable amount is fine
							corev1.ResourceName(corev1.ResourceMemory): resource.MustParse("100Mi"),
						},
					},
				},
			},
		},
	}
}

func WaitForPodToBeRunning(ctx context.Context, cli kubernetes.Interface, podNamespace, podName string, timeout time.Duration) (*corev1.Pod, error) {
	var err error
	var pod *corev1.Pod
	startTime := time.Now()
	err = k8swait.PollUntilContextTimeout(ctx, 10*time.Second, timeout, true, func(fctx context.Context) (bool, error) {
		var err2 error
		pod, err2 = cli.CoreV1().Pods(podNamespace).Get(fctx, podName, metav1.GetOptions{})
		if err2 != nil {
			return false, err2
		}
		switch pod.Status.Phase {
		case corev1.PodFailed, corev1.PodSucceeded:
			return false, fmt.Errorf("pod %q status %q which is unexpected", podName, pod.Status.Phase)
		case corev1.PodRunning:
			fmt.Fprintf(ginkgo.GinkgoWriter, "Pod %q is running! (took %v)\n", podName, time.Since(startTime))
			return true, nil
		}
		msg := fmt.Sprintf("pod %q status %q, waiting for it to be Running (with Ready = true)", podName, pod.Status.Phase)
		fmt.Fprintln(ginkgo.GinkgoWriter, msg)
		return false, nil
	})
	return pod, err
}

func ExpectPodToBeRunning(cli kubernetes.Interface, podNamespace, podName string, timeout time.Duration) *corev1.Pod {
	ginkgo.GinkgoHelper()
	pod, err := WaitForPodToBeRunning(context.Background(), cli, podNamespace, podName, timeout)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	return pod
}

func WaitPodsToBeRunningByRegex(pattern string) {
	cs, err := clientutil.New()
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	startTime := time.Now()
	gomega.EventuallyWithOffset(1, func() error {
		pods, err := GetByRegex(cs, fmt.Sprintf("%s-*", pattern))
		if err != nil {
			return err
		}
		if len(pods) == 0 {
			return fmt.Errorf("no pods found for %q", pattern)
		}

		for _, pod := range pods {
			if pod.Status.Phase != corev1.PodRunning {
				msg := fmt.Sprintf("pod %q is not in %v state", pod.Name, corev1.PodRunning)
				fmt.Println(ginkgo.GinkgoWriter, msg)
				return errors.New(msg)
			}
		}
		fmt.Fprintf(ginkgo.GinkgoWriter, "all pods running! (took %v)\n", time.Since(startTime))
		return nil
	}, 1*time.Minute, 10*time.Second).ShouldNot(gomega.HaveOccurred())
}

func GetByRegex(cs client.Client, reg string) ([]*corev1.Pod, error) {
	podNameRgx, err := regexp.Compile(reg)
	if err != nil {
		return nil, err
	}

	podList := &corev1.PodList{}
	err = cs.List(context.TODO(), podList)
	if err != nil {
		return nil, err
	}

	ret := []*corev1.Pod{}
	for _, pod := range podList.Items {
		if match := podNameRgx.FindString(pod.Name); len(match) != 0 {
			ret = append(ret, &pod)
		}
	}
	return ret, nil
}

func GetByDeployment(cli client.Client, ctx context.Context, deployment appsv1.Deployment) ([]corev1.Pod, error) {
	podList := &corev1.PodList{}
	sel, err := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
	if err != nil {
		return nil, err
	}

	err = cli.List(ctx, podList, &client.ListOptions{Namespace: deployment.Namespace, LabelSelector: sel})
	if err != nil {
		return nil, err
	}

	return podList.Items, nil
}

func GetByDaemonSet(cli client.Client, ctx context.Context, daemonset appsv1.DaemonSet) ([]corev1.Pod, error) {
	podList := &corev1.PodList{}
	sel, err := metav1.LabelSelectorAsSelector(daemonset.Spec.Selector)
	if err != nil {
		return nil, err
	}

	err = cli.List(ctx, podList, &client.ListOptions{Namespace: daemonset.Namespace, LabelSelector: sel})
	if err != nil {
		return nil, err
	}

	return podList.Items, nil
}
