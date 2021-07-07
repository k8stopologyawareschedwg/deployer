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

package utilfuncs

import (
	"context"

	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/rest"
)

func CreateCRD(config *rest.Config, crd *apiextensionv1.CustomResourceDefinition) (bool, error) {
	created := false
	kubeClient, err := apiextension.NewForConfig(config)
	if err != nil {
		return created, err
	}

	_, err = kubeClient.ApiextensionsV1().CustomResourceDefinitions().Create(context.TODO(), crd, metav1.CreateOptions{})
	if err == nil {
		created = true
	} else {
		if apierrors.IsAlreadyExists(err) {
			return created, nil
		}
		return created, err
	}

	return created, nil
}

func DeleteCRD(config *rest.Config, crd *apiextensionv1.CustomResourceDefinition) (bool, error) {
	deleted := false
	kubeClient, err := apiextension.NewForConfig(config)
	if err != nil {
		return deleted, err
	}

	_, err = kubeClient.ApiextensionsV1().CustomResourceDefinitions().Create(context.TODO(), crd.Name, metav1.DeleteOptions{})
	if err == nil {
		deleted = true
	} else {
		if apierrors.IsNotFound(err) {
			return deleted, nil
		}
		return deleted, err
	}

	return deleted, nil
}
