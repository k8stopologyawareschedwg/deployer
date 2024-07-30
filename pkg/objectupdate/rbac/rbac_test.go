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

package rbac

import (
	"reflect"
	"strings"
	"testing"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
)

func TestRoleBinding(t *testing.T) {
	testCases := []struct {
		name      string
		servAcc   string
		namespace string
		rb        *rbacv1.RoleBinding
		expected  *rbacv1.RoleBinding
	}{
		{
			name:      "namespace only",
			namespace: "rbac-test-1-updated",
			rb: &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "rbac-test-1",
					Namespace: "rbac-test-1",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      "test-kind",
						APIGroup:  "test-apigroup",
						Name:      "test-subject-name-1-1",
						Namespace: "test-subject-namespace-1-1",
					},
				},
			},
			expected: &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "rbac-test-1",
					Namespace: "rbac-test-1",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      "test-kind",
						APIGroup:  "test-apigroup",
						Name:      "test-subject-name-1-1",
						Namespace: "rbac-test-1-updated",
					},
				},
			},
		},
		{
			name:      "namespace and serviceaccount",
			servAcc:   "rbac-test-1-servacc",
			namespace: "rbac-test-1-updated",
			rb: &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "rbac-test-1",
					Namespace: "rbac-test-1",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      "test-kind",
						APIGroup:  "test-apigroup",
						Name:      "test-subject-name-1-1",
						Namespace: "test-subject-namespace-1-1",
					},
				},
			},
			expected: &rbacv1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "rbac-test-1",
					Namespace: "rbac-test-1",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      "test-kind",
						APIGroup:  "test-apigroup",
						Name:      "rbac-test-1-servacc",
						Namespace: "rbac-test-1-updated",
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.rb.DeepCopy()
			RoleBinding(got, tc.servAcc, tc.namespace)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("\ngot=%#v\nexp=%#v\n", got, tc.expected)
			}
		})
	}
}

func TestClusterRoleBinding(t *testing.T) {
	testCases := []struct {
		name      string
		servAcc   string
		namespace string
		rb        *rbacv1.ClusterRoleBinding
		expected  *rbacv1.ClusterRoleBinding
	}{
		{
			name:      "namespace only",
			namespace: "rbac-test-1-updated",
			rb: &rbacv1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "rbac-test-1",
					Namespace: "rbac-test-1",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      "test-kind",
						APIGroup:  "test-apigroup",
						Name:      "test-subject-name-1-1",
						Namespace: "test-subject-namespace-1-1",
					},
				},
			},
			expected: &rbacv1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "rbac-test-1",
					Namespace: "rbac-test-1",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      "test-kind",
						APIGroup:  "test-apigroup",
						Name:      "test-subject-name-1-1",
						Namespace: "rbac-test-1-updated",
					},
				},
			},
		},
		{
			name:      "namespace and serviceaccount",
			servAcc:   "rbac-test-1-servacc",
			namespace: "rbac-test-1-updated",
			rb: &rbacv1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "rbac-test-1",
					Namespace: "rbac-test-1",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      "test-kind",
						APIGroup:  "test-apigroup",
						Name:      "test-subject-name-1-1",
						Namespace: "test-subject-namespace-1-1",
					},
				},
			},
			expected: &rbacv1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "rbac-test-1",
					Namespace: "rbac-test-1",
				},
				Subjects: []rbacv1.Subject{
					{
						Kind:      "test-kind",
						APIGroup:  "test-apigroup",
						Name:      "rbac-test-1-servacc",
						Namespace: "rbac-test-1-updated",
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.rb.DeepCopy()
			ClusterRoleBinding(got, tc.servAcc, tc.namespace)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("\ngot=%#v\nexp=%#v\n", got, tc.expected)
			}
		})
	}
}

const testRBACBase = `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: topology-aware-scheduler-leader-elect
rules:
- apiGroups:
  - coordination.k8s.io
  resourceNames:
  - ""
  resources:
  - leases
  verbs:
  - create
- apiGroups:
  - coordination.k8s.io
  resourceNames:
  - ""
  resources:
  - leases
  verbs:
  - get
  - update
- apiGroups:
  - ""
  resourceNames:
  - ""
  resources:
  - endpoints
  verbs:
  - create
- apiGroups:
  - ""
  resourceNames:
  - ""
  resources:
  - endpoints
  verbs:
  - get
  - update`

func TestRoleForLeaderElection(t *testing.T) {
	testCases := []struct {
		desc         string
		namespace    string
		resName      string
		rbac         string
		expectedRBAC string
	}{
		{
			desc:         "empty",
			rbac:         testRBACBase,
			expectedRBAC: testRBACBase,
		},
		{
			desc:      "fix-namespace",
			namespace: "test-foobar",
			rbac:      testRBACBase,
			expectedRBAC: `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: topology-aware-scheduler-leader-elect
  namespace: test-foobar
rules:
- apiGroups:
  - coordination.k8s.io
  resourceNames:
  - ""
  resources:
  - leases
  verbs:
  - create
- apiGroups:
  - coordination.k8s.io
  resourceNames:
  - ""
  resources:
  - leases
  verbs:
  - get
  - update
- apiGroups:
  - ""
  resourceNames:
  - ""
  resources:
  - endpoints
  verbs:
  - create
- apiGroups:
  - ""
  resourceNames:
  - ""
  resources:
  - endpoints
  verbs:
  - get
  - update`,
		},
		{
			desc:    "fix-resource-name",
			resName: "test-tas-sched",
			rbac:    testRBACBase,
			expectedRBAC: `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: topology-aware-scheduler-leader-elect
rules:
- apiGroups:
  - coordination.k8s.io
  resourceNames:
  - test-tas-sched
  resources:
  - leases
  verbs:
  - create
- apiGroups:
  - coordination.k8s.io
  resourceNames:
  - test-tas-sched
  resources:
  - leases
  verbs:
  - get
  - update
- apiGroups:
  - ""
  resourceNames:
  - test-tas-sched
  resources:
  - endpoints
  verbs:
  - create
- apiGroups:
  - ""
  resourceNames:
  - test-tas-sched
  resources:
  - endpoints
  verbs:
  - get
  - update`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			obj, err := manifests.DeserializeObjectFromData([]byte(tc.rbac))
			if err != nil {
				t.Fatalf("deserialize error: %v", err)
			}
			ro, ok := obj.(*rbacv1.Role)
			if !ok {
				t.Fatalf("decoded unsupported object: %v %T", obj, obj)
			}
			RoleForLeaderElection(ro, tc.namespace, tc.resName)
			data, err := manifests.SerializeObjectToData(ro)
			if err != nil {
				t.Fatalf("serialize error: %v", err)
			}
			got := strings.TrimSpace(string(data))
			if got != tc.expectedRBAC {
				t.Errorf("got=%v\nexpected=%v\n", got, tc.expectedRBAC)
			}
		})
	}
}
