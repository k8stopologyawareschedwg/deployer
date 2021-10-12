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

	corev1 "k8s.io/api/core/v1"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	rtemanifests "github.com/k8stopologyawareschedwg/deployer/pkg/manifests/rte"
	schedmanifests "github.com/k8stopologyawareschedwg/deployer/pkg/manifests/sched"
	"github.com/k8stopologyawareschedwg/deployer/pkg/tlog"
)

type Options struct {
	Platform         platform.Platform
	WaitCompletion   bool
	Replicas         int32
	RTEConfigData    string
	PullIfNotPresent bool
}

func SetupNamespace(plat platform.Platform) (*corev1.Namespace, string, error) {
	return nil, "", fmt.Errorf("not yet implemented")
}

func Deploy(log tlog.Logger, opts Options) error {
	var err error
	log.Printf("deploying topology-aware-scheduling scheduler plugin...")

	mf, err := schedmanifests.GetManifests(opts.Platform)
	if err != nil {
		return err
	}

	rteMf, err := rtemanifests.GetManifests(opts.Platform)
	if err != nil {
		return fmt.Errorf("cannot get the rte manifests for sched: %w", err)
	}

	rteMf = rteMf.Update(rtemanifests.UpdateOptions{ConfigData: opts.RTEConfigData})
	mf = mf.Update(log, schedmanifests.UpdateOptions{
		Replicas:               opts.Replicas,
		NodeResourcesNamespace: rteMf.DaemonSet.Name,
		PullIfNotPresent:       opts.PullIfNotPresent,
	})
	log.Debugf("SCD manifests loaded")

	hp, err := deployer.NewHelper("SCD", log)
	if err != nil {
		return err
	}

	for _, wo := range mf.ToCreatableObjects(hp, log) {
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

	log.Printf("...deployed topology-aware-scheduling scheduler plugin!")
	return nil
}

func Remove(log tlog.Logger, opts Options) error {
	var err error
	log.Printf("removing topology-aware-scheduling scheduler plugin...")

	mf, err := schedmanifests.GetManifests(opts.Platform)
	if err != nil {
		return err
	}

	rteMf, err := rtemanifests.GetManifests(opts.Platform)
	if err != nil {
		return fmt.Errorf("cannot get the rte manifests for sched: %w", err)
	}

	rteMf = rteMf.Update(rtemanifests.UpdateOptions{ConfigData: opts.RTEConfigData})
	mf = mf.Update(log, schedmanifests.UpdateOptions{
		Replicas:               opts.Replicas,
		NodeResourcesNamespace: rteMf.DaemonSet.Namespace,
		PullIfNotPresent:       opts.PullIfNotPresent,
	})
	log.Debugf("SCD manifests loaded")

	hp, err := deployer.NewHelper("SCD", log)
	if err != nil {
		return err
	}

	for _, wo := range mf.ToDeletableObjects(hp, log) {
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

	log.Printf("...removed topology-aware-scheduling scheduler plugin!")
	return nil
}
