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

package sched

import (
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/fromanirh/deployer/pkg/deployer"
	schedmanifests "github.com/fromanirh/deployer/pkg/manifests/sched"
)

type Options struct {
	WaitCompletion bool
}

func Deploy(log deployer.Logger, opts Options) error {
	var err error
	log.Printf("deploying topology-aware-scheduling scheduler plugin...")

	mf, err := schedmanifests.GetManifests()
	if err != nil {
		return err
	}

	mf = mf.UpdateNamespace().UpdatePullspecs()
	log.Debugf("SCD manifests loaded")

	hp, err := deployer.NewHelper("SCD", log)
	if err != nil {
		return err
	}

	for _, obj := range mf.ToObjects() {
		if err := hp.CreateObject(obj); err != nil {
			return err
		}
	}

	if opts.WaitCompletion {
		err = waitDeploymentPodsToBeRunningByRegex(hp, log, mf.Deployment)
		if err != nil {
			return err
		}
	}

	log.Printf("...deployed topology-aware-scheduling scheduler plugin!")
	return nil
}

func waitDeploymentPodsToBeRunningByRegex(hp *deployer.Helper, log deployer.Logger, dp *appsv1.Deployment) error {
	log.Printf("wait for all the pods in deployment %s %s to be running and ready", dp.Namespace, dp.Name)
	return wait.PollImmediate(10*time.Second, 3*time.Minute, func() (bool, error) {
		pods, err := hp.GetPodsByPattern(dp.Namespace, fmt.Sprintf("%s-*", dp.Name))
		if err != nil {
			return false, err
		}
		if len(pods) == 0 {
			log.Printf("no pods found for %s %s", dp.Namespace, dp.Name)
			return false, nil
		}

		for _, pod := range pods {
			if pod.Status.Phase != corev1.PodRunning {
				log.Printf("pod %s %s not ready yet (%s)", pod.Namespace, pod.Name, pod.Status.Phase)
				return false, nil
			}
		}
		log.Printf("all the pods in deployment %s %s are running and ready!", dp.Namespace, dp.Name)
		return true, nil
	})
}

func Remove(log deployer.Logger, opts Options) error {
	var err error
	log.Printf("removing topology-aware-scheduling scheduler plugin...")

	mf, err := schedmanifests.GetManifests()
	if err != nil {
		return err
	}

	mf = mf.UpdateNamespace()
	log.Debugf("SCD manifests loaded")

	hp, err := deployer.NewHelper("SCD", log)
	if err != nil {
		return err
	}

	err = hp.DeleteObject(mf.Deployment)
	if err != nil {
		log.Printf("failed to remove: %v", err)
	} else {
		if opts.WaitCompletion {
			if err := waitDeploymentPodsToBeGoneByRegex(hp, log, mf.Deployment); err != nil {
				return err
			}
		}
	}

	err = hp.DeleteObject(mf.CRBKubernetesScheduler)
	if err != nil {
		log.Printf("failed to remove: %v", err)
	}
	err = hp.DeleteObject(mf.CRBVolumeScheduler)
	if err != nil {
		log.Printf("failed to remove: %v", err)
	}
	err = hp.DeleteObject(mf.CRBNodeResourceTopology)
	if err != nil {
		log.Printf("failed to remove: %v", err)
	}
	err = hp.DeleteObject(mf.ClusterRole)
	if err != nil {
		log.Printf("failed to remove: %v", err)
	}
	err = hp.DeleteObject(mf.RoleBinding)
	if err != nil {
		log.Printf("failed to remove: %v", err)
	}
	err = hp.DeleteObject(mf.ServiceAccount)
	if err != nil {
		log.Printf("failed to remove: %v", err)
	}
	err = hp.DeleteObject(mf.ConfigMap)
	if err != nil {
		log.Printf("failed to remove: %v", err)
	}

	log.Printf("...removed topology-aware-scheduling scheduler plugin!")
	return nil
}

func waitDeploymentPodsToBeGoneByRegex(hp *deployer.Helper, log deployer.Logger, dp *appsv1.Deployment) error {
	log.Printf("wait for all the pods in deployment %s %s to be gone", dp.Namespace, dp.Name)
	return wait.PollImmediate(10*time.Second, 3*time.Minute, func() (bool, error) {
		pods, err := hp.GetPodsByPattern(dp.Namespace, fmt.Sprintf("%s-*", dp.Name))
		if err != nil {
			return false, err
		}
		if len(pods) > 0 {
			return false, fmt.Errorf("still %d pods found for %s %s", len(pods), dp.Namespace, dp.Name)
		}
		log.Printf("all pods gone for deployment %s %s are gone!", dp.Namespace, dp.Name)
		return true, nil
	})
}
