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

package detect

import (
	"context"
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	ocpconfigv1 "github.com/openshift/api/config/v1"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
)

func TestPlatformFromLister(t *testing.T) {
	type testCase struct {
		name             string
		vers             []ocpconfigv1.ClusterVersion
		err              error
		expectedPlatform platform.Platform
		expectedError    error
	}

	unexpectedError := fmt.Errorf("unexpected error")

	testCases := []testCase{
		{
			name:             "unexpected error",
			err:              unexpectedError,
			expectedError:    unexpectedError,
			expectedPlatform: platform.Unknown,
		},
		{
			name:             "kubernetes, clusterversions not found",
			err:              errors.NewNotFound(schema.GroupResource{}, "ClusterVersions"),
			expectedPlatform: platform.Kubernetes,
		},
		{
			name:             "kubernetes, clusterversions empty",
			expectedPlatform: platform.Kubernetes,
		},
		{
			name: "openshift",
			vers: []ocpconfigv1.ClusterVersion{
				{}, // zero object is fine! We just need 1+ elements
			},
			expectedPlatform: platform.OpenShift,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cli := fakeLister{
				vers: tc.vers,
				err:  tc.err,
			}
			got, err := PlatformFromLister(context.TODO(), cli)
			if err != tc.expectedError {
				t.Errorf("got error %v expected %v", err, tc.expectedError)
			}
			if got != tc.expectedPlatform {
				t.Errorf("detect platform %v expected %v", got, tc.expectedPlatform)
			}
		})
	}
}

type fakeLister struct {
	vers []ocpconfigv1.ClusterVersion
	err  error
}

func (fake fakeLister) List(ctx context.Context, opts metav1.ListOptions) (*ocpconfigv1.ClusterVersionList, error) {
	verList := ocpconfigv1.ClusterVersionList{
		Items: fake.vers,
	}
	return &verList, fake.err
}
