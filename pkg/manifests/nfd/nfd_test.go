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

package nfd

import (
	"reflect"
	"testing"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/options"
)

const defaultNFDNamespace = "node-feature-discovery"

func TestNewWithOptions(t *testing.T) {
	type testCase struct {
		name      string
		namespace string
		mf        Manifests
		plat      platform.Platform
	}

	testCases := []testCase{
		{
			name:      "kubernetes manifests",
			namespace: "k8s-test-ns",
			plat:      platform.Kubernetes,
		},
		{
			name: "kubernetes manifests with default namespace",
			plat: platform.Kubernetes,
		},
	}

	for _, tc := range testCases {
		tc.mf, _ = NewWithOptions(options.Render{
			Platform:  tc.plat,
			Namespace: tc.namespace,
		})
		for _, obj := range tc.mf.ToObjects() {
			if tc.namespace != "" {
				if obj.GetNamespace() != "" && obj.GetNamespace() != tc.namespace {
					t.Errorf("testcase %q, NewWithOptions failed to create %q object named %q with correct namespace %q. got namespace %q instead ",
						tc.name, obj.GetObjectKind(), obj.GetName(), tc.namespace, obj.GetNamespace())
				}
			} else { // no namespace provided we should have the default
				if obj.GetNamespace() != "" && obj.GetNamespace() != defaultNFDNamespace {
					t.Errorf("testcase %q, NewWithOptions failed to create %q object named %q with default namespace %q. got namespace %q instead ",
						tc.name, obj.GetObjectKind(), obj.GetName(), defaultNFDNamespace, obj.GetNamespace())
				}
			}
		}
	}
}
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
		tc.mf, _ = NewWithOptions(options.Render{
			Platform: tc.plat,
		})
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
		tc.mf, _ = NewWithOptions(options.Render{
			Platform: tc.plat,
		})
		mfBeforeUpdate := tc.mf.Clone()
		uMf, err := tc.mf.Render(options.UpdaterDaemon{})
		if err != nil {
			t.Errorf("testcase %q, Render() failed: %v", tc.name, err)
		}

		if &uMf == &tc.mf {
			t.Errorf("testcase %q, Render() should return a pristine copy of Manifests object, thus should have different addresses", tc.name)
		}
		if !reflect.DeepEqual(mfBeforeUpdate, tc.mf) {
			t.Errorf("testcase %q, Render() should not modify the original Manifests object", tc.name)
		}
	}
}
