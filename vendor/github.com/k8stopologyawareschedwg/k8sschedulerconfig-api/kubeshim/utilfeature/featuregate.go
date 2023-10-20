/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utilfeature

import (
	"github.com/k8stopologyawareschedwg/k8sschedulerconfig-api/kubeshim/features"
)

type FeatureGateShim map[string]bool

func (fg FeatureGateShim) Enabled(name string) bool {
	return fg[name]
}

var DefaultFeatureGate FeatureGateShim

func init() {
	// keep in sync with k8s.io/kubernetes/pkg/features/kube_features.go
	DefaultFeatureGate = map[string]bool{
		features.DynamicResourceAllocation: false,
		features.PodSchedulingReadiness:    true,
		features.VolumeCapacityPriority:    false,
	}
}
