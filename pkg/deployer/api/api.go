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

package api

import (
	"log"

	"github.com/fromanirh/deployer/pkg/deployer"
	apimanifests "github.com/fromanirh/deployer/pkg/manifests/api"
)

type Options struct{}

func Deploy(logger *log.Logger, opts Options) error {
	var err error

	mf, err := apimanifests.GetManifests()
	if err != nil {
		return err
	}
	logger.Printf("  API manifests loaded")

	hp, err := deployer.NewHelper("API")
	if err != nil {
		return err
	}

	if err = hp.CreateObject(mf.Crd); err != nil {
		return err
	}

	return nil
}

func Remove(logger *log.Logger, opts Options) error {
	var err error

	mf, err := apimanifests.GetManifests()
	if err != nil {
		return err
	}
	logger.Printf("  API manifests loaded")

	hp, err := deployer.NewHelper("API")
	if err != nil {
		return err
	}

	if err = hp.DeleteObject(mf.Crd); err != nil {
		return err
	}

	return nil
}
