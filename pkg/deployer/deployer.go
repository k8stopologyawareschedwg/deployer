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
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/fromanirh/deployer/pkg/clientutil"
)

type Creator struct {
	tag string
	cli client.Client
}

func NewCreator(tag string) (*Creator, error) {
	cli, err := clientutil.New()
	if err != nil {
		return nil, err
	}
	return &Creator{
		tag: tag,
		cli: cli,
	}, nil
}

func (cr *Creator) CreateObject(obj client.Object) error {
	if err := cr.cli.Create(context.TODO(), obj); err != nil {
		return err
	}
	fmt.Printf("+%s> created %s %q\n", cr.tag, obj.GetObjectKind().GroupVersionKind().String(), obj.GetName())
	return nil
}
