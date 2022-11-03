module github.com/k8stopologyawareschedwg/deployer

go 1.18

require (
	github.com/aquasecurity/go-version v0.0.0-20210121072130-637058cfe492
	github.com/coreos/ignition/v2 v2.7.0
	github.com/google/go-cmp v0.5.6
	github.com/hashicorp/go-version v1.2.0
	github.com/k8stopologyawareschedwg/noderesourcetopology-api v0.0.12
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.17.0
	github.com/openshift/api v0.0.0-20210924154557-a4f696157341
	github.com/openshift/client-go v0.0.0-20210916133943-9acee1a0fb83
	github.com/openshift/machine-config-operator v0.0.1-0.20211105081319-76d6155c1dab
	github.com/spf13/cobra v1.4.0
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.24.7
	k8s.io/apiextensions-apiserver v0.23.0
	k8s.io/apimachinery v0.24.7
	k8s.io/client-go v0.24.7
	k8s.io/klog/v2 v2.60.1
	k8s.io/kube-scheduler v0.24.7
	k8s.io/kubelet v0.22.3
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9
	sigs.k8s.io/controller-runtime v0.11.1
	sigs.k8s.io/scheduler-plugins v0.22.7-0.20220314165158-277b6bdec18f
)

require (
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/coreos/vcontext v0.0.0-20191017033345-260217907eb5 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emicklei/go-restful v2.9.5+incompatible // indirect
	github.com/evanphx/json-patch v4.12.0+incompatible // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/go-logr/logr v1.2.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.5 // indirect
	github.com/go-openapi/swag v0.19.14 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/vincent-petithory/dataurl v0.0.0-20160330182126-9a301d65acbb // indirect
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // indirect
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8 // indirect
	golang.org/x/sys v0.0.0-20220209214540-3681064d5158 // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20220210224613-90d013bbcef8 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiserver v0.24.7 // indirect
	k8s.io/component-base v0.24.7 // indirect
	k8s.io/kube-openapi v0.0.0-20220328201542-3ee0da9b0b42 // indirect
	k8s.io/kubernetes v1.24.7 // indirect
	sigs.k8s.io/json v0.0.0-20211208200746-9f7c6b3444d2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace (
	k8s.io/api => k8s.io/api v0.24.7
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.24.7
	k8s.io/apimachinery => k8s.io/apimachinery v0.24.7
	k8s.io/apiserver => k8s.io/apiserver v0.24.7
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.24.7
	k8s.io/client-go => k8s.io/client-go v0.24.7
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.24.7
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.24.7
	k8s.io/code-generator => k8s.io/code-generator v0.24.7
	k8s.io/component-base => k8s.io/component-base v0.24.7
	k8s.io/component-helpers => k8s.io/component-helpers v0.24.7
	k8s.io/controller-manager => k8s.io/controller-manager v0.24.7
	k8s.io/cri-api => k8s.io/cri-api v0.24.7
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.24.7
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.24.7
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.24.7
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.24.7
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.24.7
	k8s.io/kubectl => k8s.io/kubectl v0.24.7
	k8s.io/kubelet => k8s.io/kubelet v0.24.7
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.24.7
	k8s.io/metrics => k8s.io/metrics v0.24.7
	k8s.io/mount-utils => k8s.io/mount-utils v0.24.7
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.24.7
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.24.7
)

replace (
	github.com/openshift/api => github.com/openshift/api v0.0.0-20211105154855-cb9596dd5fba // release-4.10
	github.com/openshift/machine-config-operator => github.com/openshift/machine-config-operator v0.0.1-0.20211105081319-76d6155c1dab // release-4.10
)
