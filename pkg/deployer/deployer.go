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

package deployer

import (
	"context"
	"regexp"

	"github.com/go-logr/logr"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/k8stopologyawareschedwg/deployer/pkg/clientutil"
)

type WaitableObject struct {
	Obj  client.Object
	Wait func() error
}

type Helper struct {
	cli client.Client
	log logr.Logger
}

func NewHelper(tag string, log logr.Logger) (*Helper, error) {
	cli, err := clientutil.New()
	if err != nil {
		return nil, err
	}
	return NewHelperWithClient(cli, tag, log), nil
}

func NewHelperWithClient(cli client.Client, tag string, log logr.Logger) *Helper {
	return &Helper{
		cli: cli,
		log: log.WithName(tag),
	}
}

func (hp *Helper) CreateObject(obj client.Object) error {
	objKind := obj.GetObjectKind().GroupVersionKind().Kind // shortcut
	if err := hp.cli.Create(context.TODO(), obj); err != nil {
		hp.log.Info("error creating", "kind", objKind, "name", obj.GetName(), "error", err)
		return err
	}
	hp.log.Info("created", "kind", objKind, "name", obj.GetName())
	return nil
}

func (hp *Helper) DeleteObject(obj client.Object) error {
	objKind := obj.GetObjectKind().GroupVersionKind().Kind // shortcut
	if err := hp.cli.Delete(context.TODO(), obj); err != nil {
		hp.log.Info("error deleting", "kind", objKind, "name", obj.GetName(), "error", err)
		return err
	}
	hp.log.Info("deleted", "kind", objKind, "name", obj.GetName())
	return nil
}

func (hp *Helper) GetObject(key client.ObjectKey, obj client.Object) error {
	return hp.cli.Get(context.TODO(), key, obj)
}

func (hp *Helper) GetPodsByPattern(namespace, pattern string) ([]*corev1.Pod, error) {
	var podList corev1.PodList
	err := hp.cli.List(context.TODO(), &podList)
	if err != nil {
		return nil, err
	}
	hp.log.Info("found matching pods", "count", len(podList.Items), "namespace", namespace, "pattern", pattern)

	podNameRgx, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	ret := []*corev1.Pod{}
	for _, pod := range podList.Items {
		if match := podNameRgx.FindString(pod.Name); len(match) != 0 {
			hp.log.Info("pod matches", "name", pod.Name)
			ret = append(ret, &pod)
		}
	}
	return ret, nil
}

func (hp *Helper) GetDaemonSetByName(namespace, name string) (*appsv1.DaemonSet, error) {
	key := client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}
	var ds appsv1.DaemonSet
	err := hp.GetObject(key, &ds)
	if err != nil {
		return nil, err
	}
	return &ds, nil
}

func (hp *Helper) IsDaemonSetRunning(namespace, name string) (bool, error) {
	log := hp.log.WithValues("namespace", namespace, "name", name)
	ds, err := hp.GetDaemonSetByName(namespace, name)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info("daemonset not found - retrying")
			return false, nil
		}
		return false, err
	}
	log.Info("daemonset", "desired", ds.Status.DesiredNumberScheduled, "current", ds.Status.CurrentNumberScheduled, "ready", ds.Status.NumberReady)
	return (ds.Status.DesiredNumberScheduled > 0 && ds.Status.DesiredNumberScheduled == ds.Status.NumberReady), nil
}

func (hp *Helper) IsDaemonSetGone(namespace, name string) (bool, error) {
	log := hp.log.WithValues("namespace", namespace, "name", name)
	ds, err := hp.GetDaemonSetByName(namespace, name)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Info("daemonset not found - gone away!")
			return true, nil
		}
		return true, err
	}
	log.Info("daemonset running", "count", ds.Status.CurrentNumberScheduled)
	return false, nil
}
