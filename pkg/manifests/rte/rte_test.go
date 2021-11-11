package rte

import (
	"strings"
	"testing"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
)

func TestMultipleUpdateWithDifferentNamespaceOptions(t *testing.T) {
	type testCase struct {
		name       string
		mf         Manifests
		options    []UpdateOptions
		expectedNs string
	}

	mfK8s, _ := GetManifests(platform.Kubernetes)
	mfOcp, _ := GetManifests(platform.OpenShift)
	testCases := []testCase{
		{
			name: "kubernetes manifests",
			mf:   mfK8s,
			options: []UpdateOptions{
				{
					Namespace: "one",
				},
				{
					Namespace: "two",
				},
			},
			expectedNs: "two",
		},
		{
			name: "openshift manifests",
			mf:   mfOcp,
			options: []UpdateOptions{
				{
					Namespace: "two",
				},
				{
					Namespace: "one",
				},
			},
			expectedNs: "one",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// with an UpdateOptions slice in size of N,
			// the mf should be updated with options that appears at the Nth index.
			for _, opt := range tc.options {
				tc.mf = tc.mf.Update(opt)
			}
			if tc.mf.DaemonSet.Namespace != tc.expectedNs {
				t.Errorf("testcase %q expected namespace %q got namespace %q", tc.name, tc.expectedNs, tc.mf.DaemonSet.Namespace)
			}
			ns := getExportedNamespace(tc.mf.DaemonSet.Spec.Template.Spec.Containers[0].Command)
			if ns != tc.expectedNs {
				t.Errorf("testcase %q expected namespace %q got namespace %q", tc.name, tc.expectedNs, ns)
			}
		})
	}
}

func getExportedNamespace(cmd []string) string {
	for _, s := range cmd {
		if strings.Contains(s, "--export-namespace=") {
			return strings.TrimPrefix(s, "--export-namespace=")
		}
	}
	return ""
}
