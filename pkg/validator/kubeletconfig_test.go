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

	"github.com/go-logr/stdr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	kubeletconfigv1beta1 "k8s.io/kubelet/config/v1beta1"
)

func TestKubeletValidations(t *testing.T) {
	nodeName := "testNode"

	type testCase struct {
		name        string
		kubeletConf *kubeletconfigv1beta1.KubeletConfiguration
		nodeVersion *version.Info
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
					Component: ComponentConfiguration,
					Setting:   "CPU",
				},
				{
					Node:      nodeName,
					Area:      AreaKubelet,
					Component: ComponentMemoryManager,
					Setting:   "policy",
				},
				{
					Node:      nodeName,
					Area:      AreaKubelet,
					Component: ComponentConfiguration,
					Setting:   "memory",
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
				MemoryManagerPolicy: ExpectedMemoryManagerPolicy,
				ReservedMemory: []kubeletconfigv1beta1.MemoryReservation{
					{
						NumaNode: 1,
					},
				},
				ReservedSystemCPUs:    "0,1",
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
				MemoryManagerPolicy: ExpectedMemoryManagerPolicy,
				ReservedMemory: []kubeletconfigv1beta1.MemoryReservation{
					{
						NumaNode: 1,
					},
				},
				ReservedSystemCPUs:    "0,1",
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
				MemoryManagerPolicy: ExpectedMemoryManagerPolicy,
				ReservedMemory: []kubeletconfigv1beta1.MemoryReservation{
					{
						NumaNode: 1,
					},
				},
				ReservedSystemCPUs: "0,1",
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
				MemoryManagerPolicy: ExpectedMemoryManagerPolicy,
				ReservedMemory: []kubeletconfigv1beta1.MemoryReservation{
					{
						NumaNode: 1,
					},
				},
				ReservedSystemCPUs:    "0,1",
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
				MemoryManagerPolicy: ExpectedMemoryManagerPolicy,
				ReservedMemory: []kubeletconfigv1beta1.MemoryReservation{
					{
						NumaNode: 1,
					},
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
				{
					Node:      nodeName,
					Area:      AreaKubelet,
					Component: ComponentConfiguration,
					Setting:   "CPU",
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
				MemoryManagerPolicy: ExpectedMemoryManagerPolicy,
				ReservedMemory: []kubeletconfigv1beta1.MemoryReservation{
					{
						NumaNode: 1,
					},
				},
				ReservedSystemCPUs:    "0,1",
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
		{
			// CAUTION: I'm not actually sure k8s <= 1.20 had all these
			// fields in the KubeletConfig, so we're bending the rules a bit here
			name: "version too old, no feature gate",
			kubeletConf: &kubeletconfigv1beta1.KubeletConfiguration{
				FeatureGates:     map[string]bool{},
				CPUManagerPolicy: ExpectedCPUManagerPolicy,
				CPUManagerReconcilePeriod: metav1.Duration{
					Duration: 5 * time.Second,
				},
				MemoryManagerPolicy: ExpectedMemoryManagerPolicy,
				ReservedMemory: []kubeletconfigv1beta1.MemoryReservation{
					{
						NumaNode: 1,
					},
				},
				ReservedSystemCPUs:    "0,1",
				TopologyManagerPolicy: ExpectedTopologyManagerPolicy,
			},
			nodeVersion: &version.Info{
				Major:      "1",
				Minor:      "20",
				GitVersion: "v1.20.5",
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
			name: "version recent enough, no feature gate",
			kubeletConf: &kubeletconfigv1beta1.KubeletConfiguration{
				FeatureGates:     map[string]bool{},
				CPUManagerPolicy: ExpectedCPUManagerPolicy,
				CPUManagerReconcilePeriod: metav1.Duration{
					Duration: 5 * time.Second,
				},
				MemoryManagerPolicy: ExpectedMemoryManagerPolicy,
				ReservedMemory: []kubeletconfigv1beta1.MemoryReservation{
					{
						NumaNode: 1,
					},
				},
				ReservedSystemCPUs:    "0,1",
				TopologyManagerPolicy: ExpectedTopologyManagerPolicy,
			},
			nodeVersion: &version.Info{
				Major:      "1",
				Minor:      "23",
				GitVersion: "v1.23.1",
			},
			expected: []ValidationResult{},
		},
	}

	vd := Validator{
		Log: stdr.New(log.New(os.Stderr, "testing ", log.LstdFlags)),
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
