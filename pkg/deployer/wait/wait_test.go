/*
 * Copyright 2023 Red Hat, Inc.
 *
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
 */

package wait

import (
	"context"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSetBaseValues(t *testing.T) {
	type testCase struct {
		name     string
		timeout  time.Duration
		interval time.Duration
		expected string
	}
	testCases := []testCase{
		{
			name:     "enforce defaults",
			interval: DefaultPollInterval,
			timeout:  DefaultPollTimeout,
			expected: "wait every 1s up to 3m0s",
		},
		{
			name:     "override interval",
			interval: 11 * time.Second,
			timeout:  DefaultPollTimeout,
			expected: "wait every 11s up to 3m0s",
		},
		{
			name:     "override timeout",
			interval: DefaultPollInterval,
			timeout:  33 * time.Second,
			expected: "wait every 1s up to 33s",
		},
		{
			name:     "override both interval and timeout",
			interval: 9 * time.Second,
			timeout:  42 * time.Second,
			expected: "wait every 9s up to 42s",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			SetBaseValues(tc.interval, tc.timeout)
			wt := With(nil, logr.Discard())
			got := wt.String()
			if got != tc.expected {
				t.Errorf("default values mismatch got [%s] expected [%s]", got, tc.expected)
			}
		})
	}
}

func TestForNamespaceDeleted(t *testing.T) {
	type testCase struct {
		name        string
		timeout     time.Duration
		interval    time.Duration
		initObjs    []client.Object
		namespace   string
		expectError bool
	}

	testCases := []testCase{
		{
			name:      "already deleted",
			timeout:   DefaultPollTimeout,
			interval:  DefaultPollInterval,
			namespace: "foobar",
		},
		{
			name:     "will never be deleted",
			timeout:  3 * time.Second,
			interval: 1 * time.Second,
			initObjs: []client.Object{
				&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "foobar",
					},
				},
			},
			namespace:   "foobar",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cli := fake.NewClientBuilder().WithObjects(tc.initObjs...).Build()

			startTime := time.Now()
			err := With(cli, testr.New(t)).Interval(tc.interval).Timeout(tc.timeout).ForNamespaceDeleted(context.TODO(), tc.namespace)
			elapsed := time.Since(startTime)

			if !tc.expectError && err != nil {
				t.Errorf("unexpected failure: %v", err)
			}
			if tc.expectError {
				if err == nil {
					t.Errorf("unexpected success")
				}
				if elapsed < tc.timeout {
					t.Errorf("terminated too early: elapsed %v timeout %v", elapsed, tc.timeout)
				}
			}
		})
	}
}
