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

package rte

import (
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/fromanirh/deployer/pkg/deployer"
	rtemanifests "github.com/fromanirh/deployer/pkg/manifests/rte"
)

type Options struct {
	WaitCompletion bool
}

func Deploy(log deployer.Logger, opts Options) error {
	var err error
	log.Printf("deploying topology-aware-scheduling topology updater...")

	mf, err := rtemanifests.GetManifests()
	if err != nil {
		return err
	}
	mf = mf.UpdateNamespace().UpdatePullspecs()
	log.Debugf("RTE manifests loaded")

	hp, err := deployer.NewHelper("RTE", log)
	if err != nil {
		return err
	}

	for _, obj := range mf.ToObjects() {
		if err := hp.CreateObject(obj); err != nil {
			return err
		}
	}

	if opts.WaitCompletion {
		if err := waitDaemonSetPodsToBeRunningByRegex(hp, log, mf.DaemonSet); err != nil {
			return err
		}
	}

	log.Printf("...deployed topology-aware-scheduling topology updater!")
	return nil
}

func waitDaemonSetPodsToBeRunningByRegex(hp *deployer.Helper, log deployer.Logger, ds *appsv1.DaemonSet) error {
	log.Printf("wait for all the pods in deployment %s %s to be running and ready", ds.Namespace, ds.Name)
	return wait.PollImmediate(1*time.Second, 3*time.Minute, func() (bool, error) {
		pods, err := hp.GetPodsByPattern(ds.Namespace, fmt.Sprintf("%s-*", ds.Name))
		if err != nil {
			return false, err
		}
		if len(pods) == 0 {
			log.Printf("no pods found for %s %s", ds.Namespace, ds.Name)
			return false, nil
		}

		for _, pod := range pods {
			if pod.Status.Phase != corev1.PodRunning {
				log.Printf("pod %s %s not ready yet (%s)", pod.Namespace, pod.Name, pod.Status.Phase)
				return false, nil
			}
		}
		log.Printf("all the pods in daemonset %s %s are running and ready!", ds.Namespace, ds.Name)
		return true, nil
	})
}

func Remove(log deployer.Logger, opts Options) error {
	var err error
	log.Printf("removing topology-aware-scheduling topology updater...")

	hp, err := deployer.NewHelper("RTE", log)
	if err != nil {
		return err
	}

	mf, err := rtemanifests.GetManifests()
	if err != nil {
		return err
	}
	mf = mf.UpdateNamespace()
	log.Debugf("RTE manifests loaded")

	// since we created everything in the namespace, we can just do
	if err := hp.DeleteObject(mf.Namespace); err != nil {
		return err
	}

	if opts.WaitCompletion {
		if err := waitDaemonSetNamespaceToBeGone(hp, log, mf.DaemonSet); err != nil {
			return err
		}
	}

	// but now let's take care of cluster-scoped resources
	err = hp.DeleteObject(mf.ClusterRoleBinding)
	if err != nil {
		log.Printf("failed to remove: %v", err)
	}
	err = hp.DeleteObject(mf.ClusterRole)
	if err != nil {
		log.Printf("failed to remove: %v", err)
	}

	log.Printf("...removed topology-aware-scheduling topology updater!")
	return nil
}

func waitDaemonSetNamespaceToBeGone(hp *deployer.Helper, log deployer.Logger, ds *appsv1.DaemonSet) error {
	log.Printf("wait for the deployment namespace %q to be gone", ds.Namespace)
	return wait.PollImmediate(1*time.Second, 3*time.Minute, func() (bool, error) {
		nsKey := types.NamespacedName{
			Name: ds.Namespace,
		}
		ns := corev1.Namespace{} // unused
		err := hp.GetObject(nsKey, &ns)
		if err == nil {
			// still present
			return false, nil
		}
		if !k8serrors.IsNotFound(err) {
			return false, err
		}
		log.Printf("namespace gone for daemonset %q!", ds.Namespace)
		return true, nil
	})
}
