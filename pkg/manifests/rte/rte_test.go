package rte

import (
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"reflect"
	"testing"
)

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
		tc.mf, _ = GetManifests(tc.plat)
		cMf := tc.mf.Clone()

		if &cMf == &tc.mf {
			t.Errorf("testcase %q, Clone() should create a pristine copy of Manifests object, thus should have different addresses", tc.name)
		}
	}
}

func TestUpdate(t *testing.T) {
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
		tc.mf, _ = GetManifests(tc.plat)
		mfBeforeUpdate := tc.mf.Clone()
		uMf := tc.mf.Update(UpdateOptions{})

		if &uMf == &tc.mf {
			t.Errorf("testcase %q, Update() should return a pristine copy of Manifests object, thus should have different addresses", tc.name)
		}
		if !reflect.DeepEqual(mfBeforeUpdate, tc.mf) {
			t.Errorf("testcase %q, Update() should not modify the original Manifests object", tc.name)
		}
	}
}
