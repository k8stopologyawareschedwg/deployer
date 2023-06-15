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
	"context"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/wait"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
	"github.com/k8stopologyawareschedwg/deployer/pkg/objectupdate"
)

const (
	DefaultSyncPeriod = 10 * time.Second
	DefaultVerbose    = 1
)

const (
	RTE string = "RTE"
	NFD string = "NFD"
)

type Options struct {
	Platform        platform.Platform
	PlatformVersion platform.Version
	WaitCompletion  bool
	RTEConfigData   string
	DaemonSet       objectupdate.DaemonSetOptions
	EnableCRIHooks  bool
}

func Deploy(env *deployer.Environment, updaterType string, opts Options) error {
	env = env.WithName(updaterType)
	env.Log.Info("deploying topology-aware-scheduling topology updater")

	ns, namespace, err := SetupNamespace(updaterType)
	if err != nil {
		return err
	}

	objs, err := getCreatableObjects(env, opts, updaterType, namespace)
	if err != nil {
		return err
	}

	env.Log.V(3).Info("manifests loaded")

	objs = append([]deployer.WaitableObject{{Obj: ns}}, objs...)

	for _, wo := range objs {
		if err := env.CreateObject(wo.Obj); err != nil {
			return err
		}
		if opts.WaitCompletion && wo.Wait != nil {
			err = wo.Wait(env.Ctx)
			if err != nil {
				return err
			}
		}
	}

	env.Log.Info("deployed topology-aware-scheduling topology updater!")
	return nil
}

func Remove(env *deployer.Environment, updaterType string, opts Options) error {
	var err error
	env = env.WithName(updaterType)
	env.Log.Info("removing topology-aware-scheduling topology updater")

	ns, err := manifests.Namespace(updaterTypeAsComponent(updaterType))
	if err != nil {
		return err
	}
	namespace := ns.Name

	objs, err := getDeletableObjects(env, opts, updaterType, namespace)
	if err != nil {
		return err
	}

	env.Log.V(3).Info("%s manifests loaded")

	objs = append(objs, deployer.WaitableObject{
		Obj:  ns,
		Wait: func(ctx context.Context) error { return wait.With(env.Cli, env.Log).ForNamespaceDeleted(ctx, ns.Name) },
	})
	for _, wo := range objs {
		err = env.DeleteObject(wo.Obj)
		if err != nil {
			continue
		}

		if !opts.WaitCompletion || wo.Wait == nil {
			continue
		}

		err = wo.Wait(env.Ctx)
		if err != nil {
			env.Log.Info("failed to wait for removal", "error", err)
		}
	}

	env.Log.Info("removed topology-aware-scheduling topology updater!")
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
