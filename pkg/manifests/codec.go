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

package manifests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	k8sjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"

	"sigs.k8s.io/controller-runtime/pkg/client"

	schedconfig "github.com/k8stopologyawareschedwg/k8sschedulerconfig-api/pkg/scheduler/apis/config"
	schedscheme "github.com/k8stopologyawareschedwg/k8sschedulerconfig-api/scheduler-plugins/apis/config/scheme"
	"github.com/k8stopologyawareschedwg/k8sschedulerconfig-api/scheduler-plugins/apis/config/v1beta2"
)

func SerializeObject(obj runtime.Object, out io.Writer) error {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	var r unstructured.Unstructured
	if err := json.Unmarshal(jsonBytes, &r.Object); err != nil {
		return err
	}

	// remove status and metadata.creationTimestamp
	unstructured.RemoveNestedField(r.Object, "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(r.Object, "template", "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(r.Object, "spec", "template", "metadata", "creationTimestamp")
	unstructured.RemoveNestedField(r.Object, "status")

	srz := k8sjson.NewYAMLSerializer(k8sjson.DefaultMetaFactory, k8sscheme.Scheme, k8sscheme.Scheme)
	return srz.Encode(&r, out)
}

func SerializeObjectToData(obj runtime.Object) ([]byte, error) {
	var buf bytes.Buffer
	err := SerializeObject(obj, &buf)
	return buf.Bytes(), err
}

func DeserializeObjectFromData(data []byte) (runtime.Object, error) {
	decode := k8sscheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(data, nil, nil)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func loadObject(path string) (runtime.Object, error) {
	data, err := src.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return DeserializeObjectFromData(data)
}

func DecodeSchedulerConfigFromData(data []byte) (*schedconfig.KubeSchedulerConfiguration, error) {
	decoder := schedscheme.Codecs.UniversalDecoder()
	obj, gvk, err := decoder.Decode(data, nil, nil)

	if err != nil {
		return nil, err
	}

	schedCfg, ok := obj.(*schedconfig.KubeSchedulerConfiguration)
	if !ok {
		return nil, fmt.Errorf("decoded unsupported type: %T gvk=%s", obj, gvk)
	}
	return schedCfg, nil
}

func EncodeSchedulerConfigToData(schedCfg *schedconfig.KubeSchedulerConfiguration) ([]byte, error) {
	yamlInfo, ok := runtime.SerializerInfoForMediaType(schedscheme.Codecs.SupportedMediaTypes(), runtime.ContentTypeYAML)
	if !ok {
		return nil, fmt.Errorf("unable to locate encoder -- %q is not a supported media type", runtime.ContentTypeYAML)
	}

	encoder := schedscheme.Codecs.EncoderForVersion(yamlInfo.Serializer, v1beta2.SchemeGroupVersion)

	var buf bytes.Buffer
	err := encoder.Encode(schedCfg, &buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func RenderObjects(objs []client.Object, w io.Writer) error {
	for _, obj := range objs {
		fmt.Fprintf(w, "---\n")
		if err := SerializeObject(obj, w); err != nil {
			return err
		}
	}

	return nil
}
