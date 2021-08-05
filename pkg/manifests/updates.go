package manifests

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/drone/envsubst"

	"github.com/fromanirh/deployer/pkg/deployer/platform"
	"github.com/fromanirh/deployer/pkg/images"
)

func UpdateRoleBinding(rb *rbacv1.RoleBinding, serviceAccount, namespace string) *rbacv1.RoleBinding {
	rb.Namespace = namespace // TODO
	for idx := 0; idx < len(rb.Subjects); idx++ {
		if serviceAccount != "" {
			rb.Subjects[idx].Name = serviceAccount
		}
		rb.Subjects[idx].Namespace = namespace
	}
	return rb
}

func UpdateClusterRoleBinding(crb *rbacv1.ClusterRoleBinding, serviceAccount, namespace string) *rbacv1.ClusterRoleBinding {
	for idx := 0; idx < len(crb.Subjects); idx++ {
		if serviceAccount != "" {
			crb.Subjects[idx].Name = serviceAccount
		}
		crb.Subjects[idx].Namespace = namespace
	}
	return crb
}

func UpdateSchedulerPluginSchedulerDeployment(dp *appsv1.Deployment) *appsv1.Deployment {
	dp.Spec.Template.Spec.Containers[0].Image = images.SchedulerPluginSchedulerImage
	return dp
}

func UpdateSchedulerPluginControllerDeployment(dp *appsv1.Deployment) *appsv1.Deployment {
	dp.Spec.Template.Spec.Containers[0].Image = images.SchedulerPluginControllerImage
	return dp
}

func UpdateResourceTopologyExporterDaemonSet(plat platform.Platform, ds *appsv1.DaemonSet, cm *corev1.ConfigMap) *appsv1.DaemonSet {
	// TODO: better match by name than assume container#0 is RTE proper (not minion)
	ds.Spec.Template.Spec.Containers[0].Image = images.ResourceTopologyExporterImage
	vars := map[string]string{
		"RTE_POLL_INTERVAL": "10s",
		"EXPORT_NAMESPACE":  ds.Namespace,
	}
	ds.Spec.Template.Spec.Containers[0].Command = UpdateResourceTopologyExporterCommand(ds.Spec.Template.Spec.Containers[0].Command, vars, plat)
	if plat == platform.OpenShift {
		// this is needed to put watches in the kubelet state dirs AND
		// to open the podresources socket in R/W mode
		if ds.Spec.Template.Spec.Containers[0].SecurityContext == nil {
			ds.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{}
		}
		ds.Spec.Template.Spec.Containers[0].SecurityContext.Privileged = newBool(true)
	}
	if cm != nil {
		ds.Spec.Template.Spec.Containers[0].VolumeMounts = append(ds.Spec.Template.Spec.Containers[0].VolumeMounts,
			corev1.VolumeMount{
				Name:      "rte-config",
				MountPath: "/etc/resource-topology-exporter/config.yaml",
			},
		)
		ds.Spec.Template.Spec.Volumes = append(ds.Spec.Template.Spec.Volumes,
			corev1.Volume{
				Name: "rte-config",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "rte-config",
						},
						Optional: newBool(true),
					},
				},
			},
		)
	}
	return ds
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
