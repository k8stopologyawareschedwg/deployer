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
 * Copyright 2022 Red Hat, Inc.
 */

package updaters

import (
	"strings"

	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/wait"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
)

const (
	RTE string = "RTE"
	NFD string = "NFD"
)

type Options struct {
	Platform         platform.Platform
	PlatformVersion  platform.Version
	WaitCompletion   bool
	PullIfNotPresent bool
	RTEConfigData    string
}

func Deploy(log_ logr.Logger, updaterType string, opts Options) error {
	log := log_.WithName(updaterType)
	log.Info("deploying topology-aware-scheduling topology updater")

	ns, namespace, err := SetupNamespace(updaterType)
	if err != nil {
		return err
	}

	hp, err := deployer.NewHelper(updaterType, log_)
	if err != nil {
		return err
	}

	objs, err := getCreatableObjects(opts, hp, log, updaterType, namespace)
	if err != nil {
		return err
	}

	log.V(3).Info("manifests loaded")

	objs = append([]deployer.WaitableObject{{Obj: ns}}, objs...)

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

	log.Info("deployed topology-aware-scheduling topology updater!")
	return nil
}

func Remove(log_ logr.Logger, updaterType string, opts Options) error {
	var err error
	log := log_.WithName(updaterType)
	log.Info("removing topology-aware-scheduling topology updater")

	hp, err := deployer.NewHelper(updaterType, log_)
	if err != nil {
		return err
	}

	ns, err := manifests.Namespace(updaterTypeAsComponent(updaterType))
	if err != nil {
		return err
	}
	namespace := ns.Name

	objs, err := getDeletableObjects(opts, hp, log, updaterType, namespace)
	if err != nil {
		return err
	}

	log.V(3).Info("%s manifests loaded")

	objs = append(objs, deployer.WaitableObject{
		Obj:  ns,
		Wait: func() error { return wait.NamespaceToBeGone(hp, log, ns.Name) },
	})
	for _, wo := range objs {
		err = hp.DeleteObject(wo.Obj)
		if err != nil {
			log.Info("failed to remove: %v", err)
			continue
		}

		if !opts.WaitCompletion || wo.Wait == nil {
			continue
		}

		err = wo.Wait()
		if err != nil {
			log.Info("failed to wait for removal", "error", err)
		}
	}

	log.Info("removed topology-aware-scheduling topology updater!")
	return nil
}

func SetupNamespace(updaterType string) (*corev1.Namespace, string, error) {
	ns, err := manifests.Namespace(updaterTypeAsComponent(updaterType))
	if err != nil {
		return nil, "", err
	}
	return ns, ns.Name, nil
}

func updaterTypeAsComponent(updaterType string) string {
	// this relation is loose, but we're validating it before use
	return strings.ToLower(updaterType)
}
