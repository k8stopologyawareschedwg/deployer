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

package validator

import (
	"log"
	"os"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
)

func TestKubeletValidations(t *testing.T) {
	nodeName := "testNode"

	type testCase struct {
		name        string
		kubeletConf *kubeletconfigv1beta1.KubeletConfiguration
		expected    []ValidationResult
	}

	testCases := []testCase{
		{
			name: "nil",
			expected: []ValidationResult{
				{
					Node:      nodeName,
					Area:      AreaKubelet,
					Component: ComponentConfiguration,
				},
			},
		},
		{
			name:        "empty",
			kubeletConf: &kubeletconfigv1beta1.KubeletConfiguration{},
			expected: []ValidationResult{
				{
					Node:      nodeName,
					Area:      AreaKubelet,
					Component: ComponentFeatureGates,
				},
				{
					Node:      nodeName,
					Area:      AreaKubelet,
					Component: ComponentCPUManager,
					Setting:   "policy",
				},
				{

					Node:      nodeName,
					Area:      AreaKubelet,
					Component: ComponentCPUManager,
					Setting:   "reconcile period",
				},
				{
					Node:      nodeName,
					Area:      AreaKubelet,
					Component: ComponentTopologyManager,
					Setting:   "policy",
				},
			},
		},
		{
			name: "correct",
			kubeletConf: &kubeletconfigv1beta1.KubeletConfiguration{
				FeatureGates: map[string]bool{
					ExpectedPodResourcesFeatureGate: true,
				},
				CPUManagerPolicy: ExpectedCPUManagerPolicy,
				CPUManagerReconcilePeriod: metav1.Duration{
					Duration: 5 * time.Second,
				},
				TopologyManagerPolicy: ExpectedTopologyManagerPolicy,
			},
			expected: []ValidationResult{},
		},
		{
			name: "missing feature gate",
			kubeletConf: &kubeletconfigv1beta1.KubeletConfiguration{
				CPUManagerPolicy: ExpectedCPUManagerPolicy,
				CPUManagerReconcilePeriod: metav1.Duration{
					Duration: 5 * time.Second,
				},
				TopologyManagerPolicy: ExpectedTopologyManagerPolicy,
			},
			expected: []ValidationResult{
				{
					Node:      nodeName,
					Area:      AreaKubelet,
					Component: ComponentFeatureGates,
				},
			},
		},
		{
			name: "missing topology manager policy",
			kubeletConf: &kubeletconfigv1beta1.KubeletConfiguration{
				FeatureGates: map[string]bool{
					ExpectedPodResourcesFeatureGate: true,
				},
				CPUManagerPolicy: ExpectedCPUManagerPolicy,
				CPUManagerReconcilePeriod: metav1.Duration{
					Duration: 5 * time.Second,
				},
			},
			expected: []ValidationResult{
				{
					Node:      nodeName,
					Area:      AreaKubelet,
					Component: ComponentTopologyManager,
					Setting:   "policy",
				},
			},
		},
		{
			name: "wrong topology manager policy",
			kubeletConf: &kubeletconfigv1beta1.KubeletConfiguration{
				FeatureGates: map[string]bool{
					ExpectedPodResourcesFeatureGate: true,
				},
				CPUManagerPolicy: ExpectedCPUManagerPolicy,
				CPUManagerReconcilePeriod: metav1.Duration{
					Duration: 5 * time.Second,
				},
				TopologyManagerPolicy: "restricted",
			},
			expected: []ValidationResult{
				{
					Node:      nodeName,
					Area:      AreaKubelet,
					Component: ComponentTopologyManager,
					Setting:   "policy",
				},
			},
		},
		{
			name: "missing cpumanager configuration",
			kubeletConf: &kubeletconfigv1beta1.KubeletConfiguration{
				FeatureGates: map[string]bool{
					ExpectedPodResourcesFeatureGate: true,
				},
				TopologyManagerPolicy: ExpectedTopologyManagerPolicy,
			},
			expected: []ValidationResult{
				{
					Node:      nodeName,
					Area:      AreaKubelet,
					Component: ComponentCPUManager,
					Setting:   "policy",
				},
				{

					Node:      nodeName,
					Area:      AreaKubelet,
					Component: ComponentCPUManager,
					Setting:   "reconcile period",
				},
			},
		},
		{
			name: "wrong cpumanager reconcile period",
			kubeletConf: &kubeletconfigv1beta1.KubeletConfiguration{
				FeatureGates: map[string]bool{
					ExpectedPodResourcesFeatureGate: true,
				},
				CPUManagerPolicy: ExpectedCPUManagerPolicy,
				CPUManagerReconcilePeriod: metav1.Duration{
					Duration: 30 * time.Second,
				},
				TopologyManagerPolicy: ExpectedTopologyManagerPolicy,
			},
			expected: []ValidationResult{
				{

					Node:      nodeName,
					Area:      AreaKubelet,
					Component: ComponentCPUManager,
					Setting:   "reconcile period",
				},
			},
		},
	}

	vd := Validator{
		Log: log.New(os.Stderr, "testing ", log.LstdFlags),
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := vd.ValidateNodeKubeletConfig(nodeName, tc.nodeVersion, tc.kubeletConf)
			if !matchValidationResults(tc.expected, got) {
				t.Fatalf("validation failed:\nexpected=%#v\ngot=%#v", tc.expected, got)
			}
		})
	}
}

func matchValidationResults(expected, got []ValidationResult) bool {
	if len(expected) != len(got) {
		return false
	}
	for _, gotvr := range got {
		if idx := findMatchingValidationResult(expected, gotvr); idx == -1 {
			return false
		}
	}
	return true
}

func findMatchingValidationResult(results []ValidationResult, desired ValidationResult) int {
	for idx, result := range results {
		if matchValidationResult(result, desired) {
			return idx
		}
	}
	return -1
}

func matchValidationResult(expected, got ValidationResult) bool {
	if expected.Node != got.Node {
		return false
	}
	if expected.Area != got.Area {
		return false
	}
	if expected.Component != got.Component {
		return false
	}
	if expected.Setting != "" && expected.Setting != "*" {
		if expected.Setting != got.Setting {
			return false
		}
	}
	return true
}
