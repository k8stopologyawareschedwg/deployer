package manifests

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/drone/envsubst"

	"github.com/fromanirh/deployer/pkg/deployer/platform"
	"github.com/fromanirh/deployer/pkg/images"
)

func UpdateSchedulerPluginDeployment(dp *appsv1.Deployment) *appsv1.Deployment {
	ret := dp.DeepCopy()
	ret.Spec.Template.Spec.Containers[0].Image = images.SchedulerPluginImage
	return ret
}

func UpdateResourceTopologyExporterDaemonSet(ds *appsv1.DaemonSet, plat platform.Platform) *appsv1.DaemonSet {
	ret := ds.DeepCopy()
	// TODO: better match by name than assume container#0 is RTE proper (not minion)
	ret.Spec.Template.Spec.Containers[0].Image = images.ResourceTopologyExporterImage
	vars := map[string]string{
		"RTE_POLL_INTERVAL": "10s",
		"EXPORT_NAMESPACE":  ds.Namespace,
	}
	ret.Spec.Template.Spec.Containers[0].Command = UpdateResourceTopologyExporterCommand(ds.Spec.Template.Spec.Containers[0].Command, vars, plat)
	if plat == platform.OpenShift {
		// this is needed to put watches in the kubelet state dirs AND
		// to open the podresources socket in R/W mode
		if ret.Spec.Template.Spec.Containers[0].SecurityContext == nil {
			ret.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{}
		}
		ret.Spec.Template.Spec.Containers[0].SecurityContext.Privileged = newBool(true)
	}
	return ret
}

func UpdateResourceTopologyExporterCommand(args []string, vars map[string]string, plat platform.Platform) []string {
	res := []string{}
	for _, arg := range args {
		newArg, err := envsubst.Eval(arg, func(key string) string { return vars[key] })
		if err != nil {
			// TODO log?
			continue
		}
		res = append(res, newArg)
	}
	if plat == platform.Kubernetes {
		res = append(res, "--kubelet-config-file=/host-var/lib/kubelet/config.yaml")
	}
	if plat == platform.OpenShift {
		// TODO
		res = append(res, "--topology-manager-policy=single-numa-node")
	}
	return res
}

func newBool(val bool) *bool {
	return &val
}
