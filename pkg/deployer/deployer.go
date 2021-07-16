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
	"log"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fromanirh/deployer/pkg/clientutil"
)

type Logger interface {
	Printf(format string, v ...interface{})
	Debugf(format string, v ...interface{})
}

type Helper struct {
	tag string
	cli client.Client
	log Logger
}

func NewHelper(tag string, log Logger) (*Helper, error) {
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
		log.Printf("-%5s> error creating %s %q: %v", hp.tag, objKind, obj.GetName(), err)
		return err
	}
	log.Printf("-%5s> created %s %q", hp.tag, objKind, obj.GetName())
	return nil
}

func (hp *Helper) DeleteObject(obj client.Object) error {
	objKind := obj.GetObjectKind().GroupVersionKind().Kind // shortcut
	if err := hp.cli.Delete(context.TODO(), obj); err != nil {
		log.Printf("-%5s> error deleting %s %q: %v", hp.tag, objKind, obj.GetName(), err)
		return err
	}
	log.Printf("-%5s> deleted %s %q", hp.tag, objKind, obj.GetName())
	return nil
}

func (hp *Helper) WaitForObjectToBeCreated(key client.ObjectKey, obj client.Object) error {
	var err error
	tries := 10
	tryInterval := 2

	for try := 0; try < tries; try++ {
		err = hp.tryGetOnce(key, obj)

		if k8serrors.IsNotFound(err) {
			time.Sleep(time.Duration(tryInterval))
		}

		if err == nil {
			return nil
		}
	}

	return err
}

func (hp *Helper) WaitForObjectToBeDeleted(key client.ObjectKey, obj client.Object) error {
	var err error
	tries := 10
	tryInterval := 2

	for try := 0; try < tries; try++ {
		err = hp.tryGetOnce(key, obj)

		if err == nil {
			time.Sleep(time.Duration(tryInterval))
			continue
		}

		if k8serrors.IsNotFound(err) {
			return nil
		}
	}

	return err
}

func (hp *Helper) tryGetOnce(key client.ObjectKey, obj client.Object) error {
	return hp.cli.Get(context.TODO(), key, obj)
}
