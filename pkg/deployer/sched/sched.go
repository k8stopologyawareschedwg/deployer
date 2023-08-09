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

	corev1 "k8s.io/api/core/v1"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	schedmanifests "github.com/k8stopologyawareschedwg/deployer/pkg/manifests/sched"
	schedwait "github.com/k8stopologyawareschedwg/deployer/pkg/objectwait/sched"
)

type Options struct {
	Platform          platform.Platform
	WaitCompletion    bool
	Replicas          int32
	RTEConfigData     string
	PullIfNotPresent  bool
	CacheResyncPeriod time.Duration
	CtrlPlaneAffinity bool
	Verbose           int
}

func SetupNamespace(plat platform.Platform) (*corev1.Namespace, string, error) {
	return nil, "", fmt.Errorf("not yet implemented")
}

func Deploy(env *deployer.Environment, opts Options) error {
	var err error
	env = env.WithName("SCD")
	env.Log.Info("deploying topology-aware-scheduling scheduler plugin")

	mf, err := schedmanifests.GetManifests(opts.Platform, "")
	if err != nil {
		return err
	}

	mf, err = mf.Render(env.Log, schedmanifests.RenderOptions{
		Replicas:          opts.Replicas,
		PullIfNotPresent:  opts.PullIfNotPresent,
		CacheResyncPeriod: opts.CacheResyncPeriod,
		CtrlPlaneAffinity: opts.CtrlPlaneAffinity,
		Verbose:           opts.Verbose,
	})
	if err != nil {
		return err
	}
	env.Log.V(3).Info("manifests loaded")

	for _, wo := range schedwait.Creatable(mf, env.Cli, env.Log) {
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

	env.Log.Info("deployed topology-aware-scheduling scheduler plugin")
	return nil
}

func Remove(env *deployer.Environment, opts Options) error {
	var err error
	env = env.WithName("SCD")
	env.Log.Info("removing topology-aware-scheduling scheduler plugin")

	mf, err := schedmanifests.GetManifests(opts.Platform, "")
	if err != nil {
		return err
	}

	mf, err = mf.Render(env.Log, schedmanifests.RenderOptions{
		Replicas:          opts.Replicas,
		PullIfNotPresent:  opts.PullIfNotPresent,
		CacheResyncPeriod: opts.CacheResyncPeriod,
		CtrlPlaneAffinity: opts.CtrlPlaneAffinity,
		Verbose:           opts.Verbose,
	})
	if err != nil {
		return err
	}
	env.Log.V(3).Info("manifests loaded")

	for _, wo := range schedwait.Deletable(mf, env.Cli, env.Log) {
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

	env.Log.Info("removed topology-aware-scheduling scheduler plugin")
	return nil
}
