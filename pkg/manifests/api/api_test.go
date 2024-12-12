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
 * Copyright 2023 Red Hat, Inc.
 */

package api

import (
	"reflect"
	"testing"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/options"
)

func TestClone(t *testing.T) {
	type testCase struct {
		name string
		mf   Manifests
		plat platform.Platform
	}

	testCases := []testCase{
		{
			name: "kubernetes manifests",
			plat: platform.Kubernetes,
		},
		{
			name: "openshift manifests",
			plat: platform.OpenShift,
		},
	}

	for _, tc := range testCases {
		var err error
		tc.mf, err = NewWithOptions(options.Render{
			Platform: tc.plat,
		})
		if err != nil {
			t.Fatalf("NewWithOptions() failed: %v", err)
		}

		cMf := tc.mf.Clone()
		if &cMf == &tc.mf {
			t.Errorf("testcase %q, Clone() should create a pristine copy of Manifests object, thus should have different addresses", tc.name)
		}
	}
}

func TestRender(t *testing.T) {
	type testCase struct {
		name string
		mf   Manifests
		plat platform.Platform
	}

	testCases := []testCase{
		{
			name: "kubernetes manifests",
			plat: platform.Kubernetes,
		},
		{
			name: "openshift manifests",
			plat: platform.OpenShift,
		},
	}

	for _, tc := range testCases {
		var err error
		tc.mf, err = NewWithOptions(options.Render{
			Platform: tc.plat,
		})
		if err != nil {
			t.Fatalf("NewWithOptions() failed: %v", err)
		}
		mfBeforeRender := tc.mf.Clone()
		uMf, err := tc.mf.Render()
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

func TestToObjects(t *testing.T) {
	type testCase struct {
		name string
		mf   Manifests
		plat platform.Platform
	}

	testCases := []testCase{
		{
			name: "kubernetes manifests",
			plat: platform.Kubernetes,
		},
		{
			name: "openshift manifests",
			plat: platform.OpenShift,
		},
	}

	for _, tc := range testCases {
		var err error
		tc.mf, err = NewWithOptions(options.Render{
			Platform: tc.plat,
		})
		if err != nil {
			t.Fatalf("NewWithOptions() failed: %v", err)
		}
		objs := tc.mf.ToObjects()
		if len(objs) == 0 {
			t.Errorf("testcase %q, ToObjects() returned zero objects", tc.name)
		}
	}
}
