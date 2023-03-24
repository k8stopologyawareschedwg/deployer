package rte

import (
	"embed"
	"path/filepath"

	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
)

const (
	NotifierName   = "rte-notifier.sh"
	ListCreateName = "rte-create.sh"
	ListDeleteName = "rte-delete.sh"
)

const (
	// OCPVersion4.11 is DEPRECATED and will be removed in the next versions
	OCPVersion411 = "v4.11"
)

const (
	selinuxPolicyDir = "selinuxpolicy"
	ocihooksDir      = "ocihooks"

	ocpVersion410 = "v4.10"
	// TODO: demote public constant here once we can remove from the public API
	ocpVersion412 = "v4.12"
	ocpVersion413 = "v4.13"
)

//go:embed ocihooks
var ocihooks embed.FS

func GetOCIHookNotifierConfig() ([]byte, error) {
	return ocihooks.ReadFile(filepath.Join(ocihooksDir, "hookconfigrtenotifier.json.template"))
}

func GetOCIHookListCreateConfig() ([]byte, error) {
	return ocihooks.ReadFile(filepath.Join(ocihooksDir, "hookconfigrtecreate.json.template"))
}

func GetOCIHookListDeleteConfig() ([]byte, error) {
	return ocihooks.ReadFile(filepath.Join(ocihooksDir, "hookconfigrtedelete.json.template"))
}

func GetOCIHookNotifier() ([]byte, error) {
	return ocihooks.ReadFile(filepath.Join(ocihooksDir, NotifierName))
}
func GetOCIHookListCreate() ([]byte, error) {
	return ocihooks.ReadFile(filepath.Join(ocihooksDir, ListCreateName))
}

func GetOCIHookListDelete() ([]byte, error) {
	return ocihooks.ReadFile(filepath.Join(ocihooksDir, ListDeleteName))
}

//go:embed selinuxpolicy
var selinuxpolicy embed.FS

//go:embed selinuxinstall.service.template
var SELinuxInstallSystemdServiceTemplate []byte

func GetSELinuxPolicy(ver platform.Version) ([]byte, error) {
	// keep it ordered from most recent supported to the oldest supported
	for _, cand := range []string{ocpVersion413, ocpVersion412, OCPVersion411, ocpVersion410} {
		// error should never happen: we control the input here
		ok, err := ver.AtLeastString(cand)
		if err != nil {
			return nil, err
		}
		if ok {
			return selinuxpolicy.ReadFile(policyPathFromVer(cand))
		}
	}
	// just in case we end up here first supported version is 4.10, hence this is a safe fallback
	return selinuxpolicy.ReadFile(policyPathFromVer(ocpVersion410))
}

func policyPathFromVer(ver string) string {
	return filepath.Join(selinuxPolicyDir, "ocp_"+ver+".cil")
}
