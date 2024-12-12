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
	"reflect"
	"testing"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/options"
)

func TestClone(t *testing.T) {
	type testCase struct {
		name        string
		mf          Manifests
		plat        platform.Platform
		platVersion platform.Version
	}

	testCases := []testCase{
		{
			name:        "kubernetes manifests",
			plat:        platform.Kubernetes,
			platVersion: platform.Version("1.23"),
		},
		{
			name:        "kubernetes manifests",
			plat:        platform.Kubernetes,
			platVersion: platform.Version("v1.24"),
		},
		{
			name:        "openshift manifests",
			plat:        platform.OpenShift,
			platVersion: platform.Version("v4.10"),
		},
		{
			name:        "openshift manifests",
			plat:        platform.OpenShift,
			platVersion: platform.Version("v4.11"),
		},
	}

	for _, tc := range testCases {
		tc.mf, _ = NewWithOptions(options.Render{
			Platform:            tc.plat,
			PlatformVersion:     tc.platVersion,
			EnableCRIHooks:      true,
			CustomSELinuxPolicy: true,
		})
		cMf := tc.mf.Clone()

		if &cMf == &tc.mf {
			t.Errorf("testcase %q, Clone() should create a pristine copy of Manifests object, thus should have different addresses", tc.name)
		}
	}
}

func TestRender(t *testing.T) {
	type testCase struct {
		name        string
		mf          Manifests
		plat        platform.Platform
		platVersion platform.Version
	}

	testCases := []testCase{
		{
			name:        "kubernetes manifests 1.23",
			plat:        platform.Kubernetes,
			platVersion: platform.Version("v1.23"),
		},
		{
			name:        "kubernetes manifests 1.24",
			plat:        platform.Kubernetes,
			platVersion: platform.Version("v1.24"),
		},
		{
			name:        "openshift manifests 4.10",
			plat:        platform.OpenShift,
			platVersion: platform.Version("v4.10"),
		},
		{
			name:        "openshift manifests 4.11",
			plat:        platform.OpenShift,
			platVersion: platform.Version("v4.11"),
		},
	}

	for _, tc := range testCases {
		tc.mf, _ = NewWithOptions(options.Render{
			Platform:            tc.plat,
			PlatformVersion:     tc.platVersion,
			EnableCRIHooks:      true,
			CustomSELinuxPolicy: true,
		})
		mfBeforeRender := tc.mf.Clone()
		uMf, err := tc.mf.Render(options.UpdaterDaemon{})
		if err != nil {
			t.Errorf("testcase %q, Render() failed: %v", tc.name, err)
		}

		if &uMf == &tc.mf {
			t.Errorf("testcase %q, Render() should return a pristine copy of Manifests object, thus should have different addresses", tc.name)
		}
		if !reflect.DeepEqual(mfBeforeRender, tc.mf) {
			t.Errorf("testcase %q, Render() should not modify the original Manifests object", tc.name)
		}
	}
}

func TestNewWithOptionsOpenShift(t *testing.T) {
	type testCase struct {
		name                    string
		plat                    platform.Platform
		platVersion             platform.Version
		withCustomSELinuxPolicy bool
	}

	testCases := []testCase{
		{
			name:                    "openshift manifests 4.10",
			plat:                    platform.OpenShift,
			platVersion:             platform.Version("v4.10"),
			withCustomSELinuxPolicy: true,
		},
		{
			name:                    "openshift manifests 4.11",
			plat:                    platform.OpenShift,
			platVersion:             platform.Version("v4.11"),
			withCustomSELinuxPolicy: true,
		},
		{
			name:        "openshift manifests 4.18",
			plat:        platform.OpenShift,
			platVersion: platform.Version("v4.18"),
		},
	}
	for _, tc := range testCases {
		mf, err := NewWithOptions(options.Render{
			Platform:            tc.plat,
			PlatformVersion:     tc.platVersion,
			Namespace:           "test",
			EnableCRIHooks:      true,
			CustomSELinuxPolicy: tc.withCustomSELinuxPolicy,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if mf.SecurityContextConstraint == nil {
			t.Fatalf("no security context constraint is generated for the OpenShift platform")
		}

		if tc.withCustomSELinuxPolicy && mf.MachineConfig == nil {
			t.Fatalf("no machine config is generated for the OpenShift platform")
		}

		if !tc.withCustomSELinuxPolicy && mf.MachineConfig != nil {
			t.Fatalf("machine config should not be generated for the OpenShift platform")
		}

		if mf.DaemonSet == nil {
			t.Fatalf("no daemon set is generated for the OpenShift platform")
		}
	}
}
