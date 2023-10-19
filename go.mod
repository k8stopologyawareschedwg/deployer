module github.com/k8stopologyawareschedwg/deployer

go 1.19

require (
	github.com/aquasecurity/go-version v0.0.0-20210121072130-637058cfe492
	github.com/coreos/ignition/v2 v2.15.0
	github.com/go-logr/logr v1.2.4
	github.com/go-logr/stdr v1.2.2
	github.com/google/go-cmp v0.5.9
	github.com/hashicorp/go-version v1.6.0
	github.com/k8stopologyawareschedwg/noderesourcetopology-api v0.1.1
	github.com/k8stopologyawareschedwg/podfingerprint v0.2.2
	github.com/onsi/ginkgo/v2 v2.11.0
	github.com/onsi/gomega v1.27.10
	github.com/openshift/api v0.0.0-20231012190404-7b36cb38c7d0
	github.com/openshift/client-go v0.0.0-20230607134213-3cd0021bbee3
	github.com/openshift/machine-config-operator v0.0.1-0.20231018130918-594da89d0bb9
	github.com/spf13/cobra v1.7.0
	github.com/spf13/pflag v1.0.6-0.20210604193023-d5e0c0615ace
	k8s.io/api v0.28.1
	k8s.io/apiextensions-apiserver v0.28.0
	k8s.io/apimachinery v0.28.1
	k8s.io/client-go v0.28.1
	k8s.io/klog/v2 v2.100.1
	k8s.io/kube-scheduler v0.27.6
	k8s.io/kubelet v0.27.6
	k8s.io/kubernetes v1.26.7
	k8s.io/utils v0.0.0-20230505201702-9f6742963106
	sigs.k8s.io/controller-runtime v0.15.2
	sigs.k8s.io/scheduler-plugins v0.26.7
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/OneOfOne/xxhash v1.2.8 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/coreos/vcontext v0.0.0-20230201181013-d72178a18687 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/emicklei/go-restful/v3 v3.9.0 // indirect
	github.com/evanphx/json-patch v5.6.0+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/pprof v0.0.0-20210720184732-4bb14d4b1be1 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.16.0 // indirect
	github.com/prometheus/client_model v0.4.0 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.10.1 // indirect
	github.com/vincent-petithory/dataurl v1.0.0 // indirect
	golang.org/x/net v0.17.0 // indirect
	golang.org/x/oauth2 v0.8.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	golang.org/x/term v0.13.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/tools v0.9.3 // indirect
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiserver v0.28.1 // indirect
	k8s.io/component-base v0.28.1 // indirect
	k8s.io/kube-openapi v0.0.0-20230525220651-2546d827e515 // indirect
	sigs.k8s.io/json v0.0.0-20221116044647-bc3834ca7abd // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)

replace (
	k8s.io/api => k8s.io/api v0.27.6
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.27.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.28.0-alpha.0.0.20231013002001-194b6f8ddf62
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20231013035634-f41beb3c1c70
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.27.6
	k8s.io/client-go => k8s.io/client-go v0.27.6
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.27.6
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.27.6
	k8s.io/code-generator => k8s.io/code-generator v0.27.6
	k8s.io/component-base => k8s.io/component-base v0.27.6
	k8s.io/component-helpers => k8s.io/component-helpers v0.27.6
	k8s.io/controller-manager => k8s.io/controller-manager v0.27.6
	k8s.io/cri-api => k8s.io/cri-api v0.27.6
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.27.6
	k8s.io/dynamic-resource-allocation => k8s.io/dynamic-resource-allocation v0.27.6
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.27.6
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.27.6
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.27.6
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.27.6
	k8s.io/kubectl => k8s.io/kubectl v0.27.6
	k8s.io/kubelet => k8s.io/kubelet v0.27.6
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.27.6
	k8s.io/metrics => k8s.io/metrics v0.27.6
	k8s.io/mount-utils => k8s.io/mount-utils v0.27.6
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.27.6
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.27.6
)

replace (
	github.com/openshift/api => github.com/openshift/api v0.0.0-20231012190404-7b36cb38c7d0 // release 4.14
	github.com/openshift/machine-config-operator => github.com/openshift/machine-config-operator v0.0.1-0.20231018130918-594da89d0bb9 // release-4.14
)
