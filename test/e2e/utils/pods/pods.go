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
	"fmt"
	"regexp"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

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

func WaitForPodToBeRunning(cli *kubernetes.Clientset, podNamespace, podName string) *corev1.Pod {
	var err error
	var pod *corev1.Pod
	gomega.EventuallyWithOffset(1, func() error {
		pod, err = cli.CoreV1().Pods(podNamespace).Get(context.TODO(), podName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		switch pod.Status.Phase {
		case v1.PodFailed, v1.PodSucceeded:
			return fmt.Errorf("pod %q status %q which is unexpected", podName, pod.Status.Phase)
		case v1.PodRunning:
			fmt.Fprintf(ginkgo.GinkgoWriter, "Pod %q is running!\n", podName)
			return nil
		}
		return fmt.Errorf("pod %q status %q, waiting for it to be Running (with Ready = true)", podName, pod.Status.Phase)
	}, 1*time.Minute, 15*time.Second).ShouldNot(gomega.HaveOccurred())
	return pod
}

func WaitPodsToBeRunningByRegex(pattern string) {
	cs, err := clientutil.New()
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

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
				return fmt.Errorf("pod %q is not in %v state", pod.Name, corev1.PodRunning)
			}
		}
		return nil
	}, 1*time.Minute, 15*time.Second).ShouldNot(gomega.HaveOccurred())
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

	ret := []*v1.Pod{}
	for _, pod := range podList.Items {
		if match := podNameRgx.FindString(pod.Name); len(match) != 0 {
			ret = append(ret, &pod)
		}
	}
	return ret, nil
}
