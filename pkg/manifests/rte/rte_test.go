package rte

import (
	"reflect"
	"testing"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
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
		tc.mf, _ = GetManifests(tc.plat, "")
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
		tc.mf, _ = GetManifests(tc.plat, "")
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

func TestGetManifestsOpenShift(t *testing.T) {
	mf, err := GetManifests(platform.OpenShift, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mf.SecurityContextConstraint == nil {
		t.Fatalf("no security context constraint is generated for the OpenShift platform")
	}

	if mf.MachineConfig == nil {
		t.Fatalf("no machine config is generated for the OpenShift platform")
	}

	if mf.DaemonSet == nil {
		t.Fatalf("no daemon set is generated for the OpenShift platform")
	}
}
