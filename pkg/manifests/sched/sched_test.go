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
	"reflect"
	"testing"

	"github.com/go-logr/logr/testr"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
)

func TestClone(t *testing.T) {
	type testCase struct {
		name string
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
		mf, _ := GetManifests(tc.plat, "")
		cMf := mf.Clone()

		if &cMf == &mf {
			t.Errorf("testcase %q, Clone() should create a pristine copy of Manifests object, thus should have different addresses", tc.name)
		}
	}
}

func TestRender(t *testing.T) {
	type testCase struct {
		name string
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
		mf, _ := GetManifests(tc.plat, "")
		mfBeforeRender := mf.Clone()
		uMf, err := mf.Render(testr.New(t), RenderOptions{
			Replicas: int32(1),
		})
		if err != nil {
			t.Errorf("testcase %q, Render() failed: %v", tc.name, err)
		}

		if &uMf == &mf {
			t.Errorf("testcase %q, Render() should return a pristine copy of Manifests object, thus should have different addresses", tc.name)
		}
		if !reflect.DeepEqual(mfBeforeRender, mf) {
			t.Errorf("testcase %q, Render() should not modify the original Manifests object", tc.name)
		}
	}
}
