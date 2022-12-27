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

package wait

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

func PodsToBeRunningByRegex(cli client.Client, log logr.Logger, namespace, name string) error {
	log = log.WithValues("namespace", namespace, "name", name)
	log.Info("wait for all the pods in group to be running and ready")
	return wait.PollImmediate(1*time.Second, 3*time.Minute, func() (bool, error) {
		pods, err := getPodsByPattern(cli, log, namespace, fmt.Sprintf("%s-*", name))
		if err != nil {
			return false, err
		}
		if len(pods) == 0 {
			log.Info("no pods found")
			return false, nil
		}

		for _, pod := range pods {
			if pod.Status.Phase != corev1.PodRunning {
				log.Info("pod not ready yet", "podNamespace", pod.Namespace, "podName", pod.Name, "podPhase", pod.Status.Phase)
				return false, nil
			}
		}
		log.Info("all the pods in daemonset are running and ready!")
		return true, nil
	})
}

func PodsToBeGoneByRegex(cli client.Client, log logr.Logger, namespace, name string) error {
	log = log.WithValues("namespace", namespace, "name", name)
	log.Info("wait for all the pods in deployment to be gone")
	return wait.PollImmediate(10*time.Second, 3*time.Minute, func() (bool, error) {
		pods, err := getPodsByPattern(cli, log, namespace, fmt.Sprintf("%s-*", name))
		if err != nil {
			return false, err
		}
		if len(pods) > 0 {
			return false, fmt.Errorf("still %d pods found for %s %s", len(pods), namespace, name)
		}
		log.Info("all pods gone for deployment %s %s are gone!")
		return true, nil
	})
}

func NamespaceToBeGone(cli client.Client, log logr.Logger, namespace string) error {
	log = log.WithValues("namespace", namespace)
	log.Info("wait for the namespace to be gone")
	return wait.PollImmediate(1*time.Second, 3*time.Minute, func() (bool, error) {
		nsKey := types.NamespacedName{
			Name: namespace,
		}
		ns := corev1.Namespace{} // unused
		err := cli.Get(context.TODO(), nsKey, &ns)
		if err == nil {
			// still present
			return false, nil
		}
		if !k8serrors.IsNotFound(err) {
			return false, err
		}
		log.Info("namespace gone!")
		return true, nil
	})
}

func DaemonSetToBeRunning(cli client.Client, log_ logr.Logger, namespace, name string) error {
	log := log_.WithValues("namespace", namespace, "name", name)
	log.Info("wait for the daemonset to be running")
	return wait.PollImmediate(3*time.Second, 3*time.Minute, func() (bool, error) {
		return isDaemonSetRunning(cli, log, namespace, name)
	})
}

func DaemonSetToBeGone(cli client.Client, log_ logr.Logger, namespace, name string) error {
	log := log_.WithValues("namespace", namespace, "name", name)
	log.Info("wait for the daemonset to be gone")
	return wait.PollImmediate(3*time.Second, 3*time.Minute, func() (bool, error) {
		return isDaemonSetGone(cli, log, namespace, name)
	})
}

func isDaemonSetRunning(cli client.Client, log logr.Logger, namespace, name string) (bool, error) {
	ds, err := getDaemonSetByName(cli, namespace, name)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info("daemonset not found - retrying")
			return false, nil
		}
		return false, err
	}
	log.Info("daemonset", "desired", ds.Status.DesiredNumberScheduled, "current", ds.Status.CurrentNumberScheduled, "ready", ds.Status.NumberReady)
	return (ds.Status.DesiredNumberScheduled > 0 && ds.Status.DesiredNumberScheduled == ds.Status.NumberReady), nil
}

func isDaemonSetGone(cli client.Client, log logr.Logger, namespace, name string) (bool, error) {
	ds, err := getDaemonSetByName(cli, namespace, name)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info("daemonset not found - gone away!")
			return true, nil
		}
		return true, err
	}
	log.Info("daemonset running", "count", ds.Status.CurrentNumberScheduled)
	return false, nil
}

func getDaemonSetByName(cli client.Client, namespace, name string) (*appsv1.DaemonSet, error) {
	key := client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}
	var ds appsv1.DaemonSet
	err := cli.Get(context.TODO(), key, &ds)
	return &ds, err
}

func getPodsByPattern(cli client.Client, log logr.Logger, namespace, pattern string) ([]*corev1.Pod, error) {
	var podList corev1.PodList
	err := cli.List(context.TODO(), &podList)
	if err != nil {
		return nil, err
	}
	log.Info("found matching pods", "count", len(podList.Items), "namespace", namespace, "pattern", pattern)

	podNameRgx, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	ret := []*corev1.Pod{}
	for _, pod := range podList.Items {
		if match := podNameRgx.FindString(pod.Name); len(match) != 0 {
			log.Info("pod matches", "name", pod.Name)
			ret = append(ret, &pod)
		}
	}
	return ret, nil
}
