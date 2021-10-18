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

	corev1 "k8s.io/api/core/v1"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/wait"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
	rtemanifests "github.com/k8stopologyawareschedwg/deployer/pkg/manifests/rte"
	"github.com/k8stopologyawareschedwg/deployer/pkg/tlog"
)

type Options struct {
	Platform         platform.Platform
	WaitCompletion   bool
	RTEConfigData    string
	PullIfNotPresent bool
}

func SetupNamespace(plat platform.Platform) (*corev1.Namespace, string, error) {
	if plat == platform.Kubernetes {
		ns, err := manifests.Namespace(manifests.ComponentResourceTopologyExporter)
		if err != nil {
			return nil, "", err
		}
		return ns, ns.Name, nil
	}
	if plat == platform.OpenShift {
		return nil, rtemanifests.NamespaceOpenShift, nil
	}
	return nil, "", fmt.Errorf("unsupported platform: %q", plat)
}

func Deploy(log tlog.Logger, opts Options) error {
	log.Printf("deploying topology-aware-scheduling topology updater...")

	ns, namespace, err := SetupNamespace(opts.Platform)
	if err != nil {
		return err
	}

	mf, err := rtemanifests.GetManifests(opts.Platform)
	if err != nil {
		return err
	}
	mf = mf.Update(rtemanifests.UpdateOptions{
		ConfigData:       opts.RTEConfigData,
		PullIfNotPresent: opts.PullIfNotPresent,
		Namespace:        namespace,
	})
	log.Debugf("RTE manifests loaded")

	hp, err := deployer.NewHelper("RTE", log)
	if err != nil {
		return err
	}

	objs := mf.ToCreatableObjects(hp, log)
	if opts.Platform == platform.Kubernetes {
		objs = append([]deployer.WaitableObject{{Obj: ns}}, objs...)
	}
	for _, wo := range objs {
		if err := hp.CreateObject(wo.Obj); err != nil {
			return err
		}
		if opts.WaitCompletion && wo.Wait != nil {
			err = wo.Wait()
			if err != nil {
				return err
			}
		}
	}

	log.Printf("...deployed topology-aware-scheduling topology updater!")
	return nil
}

func Remove(log tlog.Logger, opts Options) error {
	var err error
	log.Printf("removing topology-aware-scheduling topology updater...")

	hp, err := deployer.NewHelper("RTE", log)
	if err != nil {
		return err
	}

	ns, err := manifests.Namespace(manifests.ComponentResourceTopologyExporter)
	if err != nil {
		return err
	}
	namespace := ""
	if opts.Platform == platform.Kubernetes {
		namespace = ns.Name
	}
	if opts.Platform == platform.OpenShift {
		namespace = rtemanifests.NamespaceOpenShift
	}

	mf, err := rtemanifests.GetManifests(opts.Platform)
	if err != nil {
		return err
	}
	mf = mf.Update(rtemanifests.UpdateOptions{
		ConfigData:       opts.RTEConfigData,
		PullIfNotPresent: opts.PullIfNotPresent,
		Namespace:        namespace,
	})
	log.Debugf("RTE manifests loaded")

	objs := mf.ToDeletableObjects(hp, log)
	if opts.Platform == platform.Kubernetes {
		objs = append(objs, deployer.WaitableObject{
			Obj:  ns,
			Wait: func() error { return wait.NamespaceToBeGone(hp, log, ns.Name) },
		})
	}
	for _, wo := range objs {
		err = hp.DeleteObject(wo.Obj)
		if err != nil {
			log.Printf("failed to remove: %v", err)
			continue
		}

		if !opts.WaitCompletion || wo.Wait == nil {
			continue
		}

		err = wo.Wait()
		if err != nil {
			log.Printf("failed to wait for removal: %v", err)
		}
	}

	log.Printf("...removed topology-aware-scheduling topology updater!")
	return nil
}
