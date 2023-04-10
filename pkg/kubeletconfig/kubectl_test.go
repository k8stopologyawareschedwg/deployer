/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubeletconfig

import (
	"reflect"
	"testing"

	"github.com/go-logr/logr/testr"
)

func TestKubectlBase(t *testing.T) {
	kc := NewKubectl(testr.New(t), "/bin/kubectl", "/home/test/kubeconfig").WithAPIServer("127.0.0.1:12345").WithNamespace("foobar")
	args := kc.Arguments("testing")
	expectedArgs := []string{"--kubeconfig=/home/test/kubeconfig", "--server=127.0.0.1:12345", "--namespace=foobar", "testing"}
	if !reflect.DeepEqual(args, expectedArgs) {
		t.Errorf("arguments: got=%#v expected=%#v", args, expectedArgs)
	}
}
