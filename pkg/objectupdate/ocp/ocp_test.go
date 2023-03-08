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

package ocp

import (
	"testing"

	securityv1 "github.com/openshift/api/security/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSecurityContextConstraintsAppend(t *testing.T) {
	scc := securityv1.SecurityContextConstraints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "scc-test-name-1",
			Namespace: "scc-test-ns-1",
		},
		Users: []string{"foo", "bar"},
	}
	sa := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sa-test-name-1",
			Namespace: "sa-test-ns-1",
		},
	}
	saName := MakeSecurityContextConstraintName(sa)

	SecurityContextConstraint(&scc, &sa)
	if !isIncluded(scc.Users, saName) {
		t.Errorf("%q not added", saName)
	}
}

func TestSecurityContextConstraintsNoDupes(t *testing.T) {
	sa := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sa-test-name-1",
			Namespace: "sa-test-ns-1",
		},
	}
	saName := MakeSecurityContextConstraintName(sa)

	scc := securityv1.SecurityContextConstraints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "scc-test-name-1",
			Namespace: "scc-test-ns-1",
		},
		Users: []string{"foo", saName, "bar"},
	}

	SecurityContextConstraint(&scc, &sa)
	if countOccurrences(scc.Users, saName) > 1 {
		t.Errorf("%q added multiple times", saName)
	}
}

func countOccurrences(strs []string, st string) int {
	count := 0
	for _, str := range strs {
		if str == st {
			count++
		}
	}
	return count
}
func isIncluded(strs []string, st string) bool {
	return countOccurrences(strs, st) > 0
}
