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

package updater

import (
	"fmt"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/tlog"
)

type Type string

const (
	RTE Type = "RTE"
	NFD Type = "NFD"
)

type Options struct {
	Platform          platform.Platform
	WaitCompletion    bool
	UpdaterConfigData string
	PullIfNotPresent  bool
	UpdaterType       string
}

func Deploy(log tlog.Logger, opts Options) error {
	switch Type(opts.UpdaterType) {
	case RTE:
		return deployRTE(log, rteOptions{
			Options:       opts,
			RTEConfigData: opts.UpdaterConfigData,
		})
	case NFD:
		return deployNFD(log, nfdOptions{
			Options:       opts,
			NFDConfigData: opts.UpdaterConfigData,
		})
	}
	return fmt.Errorf("%s is invalid updater type", opts.UpdaterType)
}

func Remove(log tlog.Logger, opts Options) error {
	switch Type(opts.UpdaterType) {
	case RTE:
		return removeRTE(log, rteOptions{
			Options:       opts,
			RTEConfigData: opts.UpdaterConfigData,
		})
	case NFD:
		return removeNFD(log, nfdOptions{
			Options:       opts,
			NFDConfigData: opts.UpdaterConfigData,
		})
	}
	return fmt.Errorf("%s is invalid updater type", opts.UpdaterType)
}
