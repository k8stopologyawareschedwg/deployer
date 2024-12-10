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

package dump

import (
	"context"

	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/k8stopologyawareschedwg/deployer/pkg/clientutil"
	"github.com/k8stopologyawareschedwg/deployer/pkg/deployer/platform"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests/rte"
	"github.com/k8stopologyawareschedwg/deployer/pkg/manifests/sched"
	"github.com/k8stopologyawareschedwg/deployer/pkg/options"
	e2epods "github.com/k8stopologyawareschedwg/deployer/test/e2e/utils/pods"
)

// todo: port to deployer.Environment

func ClusterState(ctx context.Context, cli client.Client, updatedPod *corev1.Pod) error {
	if err := ResourceTopologyExporterPods(ctx, cli); err != nil {
		return err
	}
	if err := SchedulerPods(ctx, cli); err != nil {
		return err
	}
	if err := WorkloadPods(ctx, updatedPod); err != nil {
		return err
	}
	return nil
}

func SchedulerPods(ctx context.Context, cli client.Client) error {
	ns, err := manifests.Namespace(manifests.ComponentSchedulerPlugin)
	if err != nil {
		return err
	}

	// TODO: autodetect the platform
	mfs, err := sched.GetManifests(platform.Kubernetes, ns.Name)
	if err != nil {
		return err
	}
	mfs, err = mfs.Render(logr.Discard(), options.Scheduler{
		Replicas: int32(1),
	})
	if err != nil {
		return err
	}

	k8sCli, err := clientutil.NewK8s()
	if err != nil {
		return err
	}

	pods, err := e2epods.GetByDeployment(cli, ctx, *mfs.DPScheduler)
	if err != nil {
		return err
	}

	klog.Warning(">>> scheduler pod status begin:\n")
	for idx := range pods {
		pod := pods[idx].DeepCopy()
		pod.ManagedFields = nil

		data, err := yaml.Marshal(pod)
		if err != nil {
			return err
		}

		klog.Warningf("%s\n---\n", string(data))

		e2epods.LogEventsForPod(k8sCli, ctx, pod.Namespace, pod.Name)
		klog.Warningf("---\n")
	}

	var cm corev1.ConfigMap
	key := client.ObjectKey{
		Namespace: "tas-scheduler",
		Name:      "scheduler-config",
	}
	err = cli.Get(ctx, key, &cm)
	if err == nil {
		// skip errors until we can autodetect the CM key
		klog.Infof("scheduler config:\n%s", cm.Data["scheduler-config.yaml"])
	} else {
		klog.Errorf("error getting config map: %v")
	}

	klog.Warning(">>> scheduler pod status end\n")
	return nil
}

func WorkloadPods(ctx context.Context, pod *corev1.Pod) error {
	pod = pod.DeepCopy()

	k8sCli, err := clientutil.NewK8s()
	if err != nil {
		return err
	}

	klog.Warning(">>> workload pod status begin:\n")
	pod.ManagedFields = nil

	data, err := yaml.Marshal(pod)
	if err != nil {
		return err
	}

	klog.Warningf("%s\n---\n", string(data))

	e2epods.LogEventsForPod(k8sCli, ctx, pod.Namespace, pod.Name)
	klog.Warningf("---\n")
	klog.Warning(">>> workload pod status end\n")
	return nil
}

func ResourceTopologyExporterPods(ctx context.Context, cli client.Client) error {
	ns, err := manifests.Namespace(manifests.ComponentResourceTopologyExporter)
	if err != nil {
		return err
	}

	// TODO: autodetect the platform
	mfs, err := rte.GetManifests(platform.Kubernetes, platform.Version("1.23"), ns.Name, true)
	if err != nil {
		return err
	}
	mfs, err = mfs.Render(options.UpdaterDaemon{Namespace: ns.Name})
	if err != nil {
		return err
	}

	k8sCli, err := clientutil.NewK8s()
	if err != nil {
		return err
	}

	pods, err := e2epods.GetByDaemonSet(cli, ctx, *mfs.DaemonSet)
	if err != nil {
		return err
	}

	klog.Warning(">>> RTE pod status begin:\n")
	if len(pods) > 1 {
		klog.Warningf("UNEXPECTED POD COUNT %d: dumping only the first", len(pods))
	}
	if len(pods) > 0 {
		pod := pods[0].DeepCopy()
		pod.ManagedFields = nil

		logs, err := e2epods.GetLogsForPod(k8sCli, pod.Namespace, pod.Name, pod.Spec.Containers[0].Name)
		if err == nil {
			// skip errors until we can autodetect the CM key
			klog.Infof(">>> RTE logs begin:\n%s\n>>> RTE logs end", logs)
		} else {
			klog.Errorf("error getting logs for pod %s/%s/%s", pod.Namespace, pod.Name, pod.Spec.Containers[0].Name)
		}
	}

	klog.Warning(">>> RTE pod status end\n")
	return nil
}
