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
	"github.com/fromanirh/deployer/pkg/deployer"
	"github.com/fromanirh/deployer/pkg/deployer/platform"
	schedmanifests "github.com/fromanirh/deployer/pkg/manifests/sched"
)

type Options struct {
	Platform       platform.Platform
	WaitCompletion bool
}

func Deploy(log deployer.Logger, opts Options) error {
	var err error
	log.Printf("deploying topology-aware-scheduling scheduler plugin...")

	mf, err := schedmanifests.GetManifests(opts.Platform)
	if err != nil {
		return err
	}

	mf = mf.Update()
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

func Remove(log deployer.Logger, opts Options) error {
	var err error
	log.Printf("removing topology-aware-scheduling scheduler plugin...")

	mf, err := schedmanifests.GetManifests(opts.Platform)
	if err != nil {
		return err
	}

	mf = mf.Update()
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
