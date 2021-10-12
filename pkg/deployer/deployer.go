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

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/k8stopologyawareschedwg/deployer/pkg/clientutil"
	"github.com/k8stopologyawareschedwg/deployer/pkg/tlog"
)

type WaitableObject struct {
	Obj  client.Object
	Wait func() error
}

type Helper struct {
	tag string
	cli client.Client
	log tlog.Logger
}

func NewHelper(tag string, log tlog.Logger) (*Helper, error) {
	cli, err := clientutil.New()
	if err != nil {
		return nil, err
	}
	return &Helper{
		tag: tag,
		cli: cli,
		log: log,
	}, nil
}

func (hp *Helper) CreateObject(obj client.Object) error {
	objKind := obj.GetObjectKind().GroupVersionKind().Kind // shortcut
	if err := hp.cli.Create(context.TODO(), obj); err != nil {
		hp.log.Printf("-%5s> error creating %s %q: %v", hp.tag, objKind, obj.GetName(), err)
		return err
	}
	hp.log.Printf("-%5s> created %s %q", hp.tag, objKind, obj.GetName())
	return nil
}

func (hp *Helper) DeleteObject(obj client.Object) error {
	objKind := obj.GetObjectKind().GroupVersionKind().Kind // shortcut
	if err := hp.cli.Delete(context.TODO(), obj); err != nil {
		hp.log.Printf("-%5s> error deleting %s %q: %v", hp.tag, objKind, obj.GetName(), err)
		return err
	}
	hp.log.Printf("-%5s> deleted %s %q", hp.tag, objKind, obj.GetName())
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
	hp.log.Debugf("found %d pods in namespace %q matching pattern %q", len(podList.Items), namespace, pattern)

	podNameRgx, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	ret := []*corev1.Pod{}
	for _, pod := range podList.Items {
		if match := podNameRgx.FindString(pod.Name); len(match) != 0 {
			hp.log.Debugf("pod %q matches", pod.Name)
			ret = append(ret, &pod)
		}
	}
	return ret, nil
}
