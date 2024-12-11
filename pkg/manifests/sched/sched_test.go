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
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
	"github.com/k8stopologyawareschedwg/deployer/pkg/options"
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
		t.Run(tc.name, func(t *testing.T) {
			mf, _ := NewWithOptions(options.Render{
				Platform: tc.plat,
			})
			cMf := mf.Clone()

			if &cMf == &mf {
				t.Errorf("testcase %q, Clone() should create a pristine copy of Manifests object, thus should have different addresses", tc.name)
			}
		})
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
		t.Run(tc.name, func(t *testing.T) {
			mf, err := NewWithOptions(options.Render{
				Platform: tc.plat,
			})
			if err != nil {
				t.Errorf("testcase %q, NewWithOptions(platform=%s) failed: %v", tc.name, tc.plat, err)
			}

			mfBeforeRender := mf.Clone()
			uMf, err := mf.Render(testr.New(t), options.Scheduler{
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
		})
	}
}

// TODO: stopgap until we have good render coverage for these cases. We will need a lot of work and love in TestRender for this.
func Test_leaderElectionParamsFromOpts(t *testing.T) {
	type testCase struct {
		name           string
		opts           options.Scheduler
		expectedParams manifests.LeaderElectionParams
		expectedOK     bool
		expectedError  error
	}

	testCases := []testCase{
		{
			name: "all zeros",
		},
		{
			name: "only flag set",
			opts: options.Scheduler{
				LeaderElection: true,
			},
			expectedOK: true,
			expectedParams: manifests.LeaderElectionParams{
				LeaderElect:       true,
				ResourceName:      manifests.LeaderElectionDefaultName,
				ResourceNamespace: manifests.LeaderElectionDefaultNamespace,
			},
		},
		{
			name: "resource non namespaced, missing sep",
			opts: options.Scheduler{
				LeaderElection:         true,
				LeaderElectionResource: "foobar",
			},
			expectedOK: true,
			expectedParams: manifests.LeaderElectionParams{
				LeaderElect:       true,
				ResourceName:      "foobar",
				ResourceNamespace: manifests.LeaderElectionDefaultNamespace,
			},
		},
		{
			name: "empty namespace",
			opts: options.Scheduler{
				LeaderElection:         true,
				LeaderElectionResource: "/foobar",
			},
			expectedOK: true,
			expectedParams: manifests.LeaderElectionParams{
				LeaderElect:       true,
				ResourceName:      "foobar",
				ResourceNamespace: manifests.LeaderElectionDefaultNamespace,
			},
		},
		{
			name: "empty names",
			opts: options.Scheduler{
				LeaderElection:         true,
				LeaderElectionResource: "foobar/",
			},
			expectedOK: true,
			expectedParams: manifests.LeaderElectionParams{
				LeaderElect:       true,
				ResourceNamespace: "foobar",
				ResourceName:      manifests.LeaderElectionDefaultName,
			},
		},
		{
			name: "namespace and name",
			opts: options.Scheduler{
				LeaderElection:         true,
				LeaderElectionResource: "foo/bar",
			},
			expectedOK: true,
			expectedParams: manifests.LeaderElectionParams{
				LeaderElect:       true,
				ResourceNamespace: "foo",
				ResourceName:      "bar",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok, err := leaderElectionParamsFromOpts(tc.opts)
			if (err != nil) != (tc.expectedError != nil) {
				t.Fatalf("got error %v expected error %v", err, tc.expectedError)
			}
			if ok != tc.expectedOK {
				t.Errorf("got ok %v expected %v", ok, tc.expectedOK)
			}
			if !reflect.DeepEqual(got, tc.expectedParams) {
				t.Errorf("got params %v expected %v", got, tc.expectedParams)
			}
		})
	}
}
