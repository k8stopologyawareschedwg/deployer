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

package e2e

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/k8stopologyawareschedwg/deployer/pkg/validator"
)

var (
	deployerBaseDir string
	binariesPath    string
)

func TestE2E(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Deployer Suite")
}

var _ = ginkgo.BeforeSuite(func() {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		ginkgo.Fail("Cannot retrieve tests directory")
	}
	basedir := filepath.Dir(file)
	deployerBaseDir = filepath.Clean(filepath.Join(basedir, "..", ".."))
	binariesPath = filepath.Clean(filepath.Join(deployerBaseDir, "_out"))
	fmt.Fprintf(ginkgo.GinkgoWriter, "using binaries at %q\n", binariesPath)
})

type validationOutput struct {
	Success bool                         `json:"success"`
	Errors  []validator.ValidationResult `json:"errors,omitempty"`
}

func (vo validationOutput) String() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "validation: success=%t\n", vo.Success)
	for _, vErr := range vo.Errors {
		fmt.Fprintf(&sb, "validation: error: %s\n", vErr.String())
	}
	return sb.String()
}

type imageOutput struct {
	TopologyUpdater     string `json:"topology_updater"`
	SchedulerPlugin     string `json:"scheduler_plugin"`
	SchedulerController string `json:"scheduler_controller"`
}

func waitForReasource(body interface{}) {
	gomega.Eventually(body)
}
