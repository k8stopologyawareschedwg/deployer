/*
Copyright 2019 The Kubernetes Authors.

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
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-logr/logr"

	"k8s.io/client-go/tools/clientcmd"
)

const (
	DefaultKubectlPath = "/bin/kubectl"
)

type Kubectl struct {
	logger      logr.Logger
	kubectlPath string
	kubeConfig  string
	apiserver   string
	namespace   string
}

func NewKubectl(logger logr.Logger, kubectlPath, kubeConfig string) *Kubectl {
	return &Kubectl{
		logger:      logger,
		kubectlPath: kubectlPath,
		kubeConfig:  kubeConfig,
	}
}

func NewKubectlFromEnv(logger logr.Logger) *Kubectl {
	kubeConfig, ok := os.LookupEnv("KUBECONFIG")
	if !ok {
		kubeConfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
		logger.Info("using default kubeconfig", "path", kubeConfig)
	}
	kubectlPath, ok := os.LookupEnv("KUBECTL")
	if !ok {
		var err error
		kubectlPath, err = exec.LookPath("kubectl")
		if err != nil {
			logger.Info("kubectl not found, falling back to hardcoded default", "error", err)
			kubectlPath = DefaultKubectlPath
		}
		logger.Info("using kubectl", "path", kubectlPath)
	}
	return NewKubectl(logger, kubectlPath, kubeConfig)
}

func (kc *Kubectl) WithAPIServer(apiserver string) *Kubectl {
	return &Kubectl{
		kubectlPath: kc.kubectlPath,
		kubeConfig:  kc.kubeConfig,
		namespace:   kc.namespace,
		apiserver:   apiserver,
	}
}

func (kc *Kubectl) WithNamespace(namespace string) *Kubectl {
	return &Kubectl{
		kubectlPath: kc.kubectlPath,
		kubeConfig:  kc.kubeConfig,
		apiserver:   kc.apiserver,
		namespace:   namespace,
	}
}

func (kc *Kubectl) IsReady() (bool, error) {
	if _, err := os.Stat(kc.kubeConfig); err != nil {
		return false, fmt.Errorf("invalid kubeconfig: %w", err)
	}
	if _, err := os.Stat(kc.kubectlPath); err != nil {
		return false, fmt.Errorf("invalid kubectl: %w", err)
	}
	return true, nil
}

func (kc *Kubectl) Arguments(args ...string) []string {
	defaultArgs := []string{
		fmt.Sprintf("--%s=%s", clientcmd.RecommendedConfigPathFlag, kc.kubeConfig),
	}
	if kc.apiserver != "" {
		defaultArgs = append(defaultArgs, fmt.Sprintf("--%s=%s", clientcmd.FlagAPIServer, kc.apiserver))
	}
	if kc.namespace != "" {
		defaultArgs = append(defaultArgs, fmt.Sprintf("--namespace=%s", kc.namespace))
	}
	return append(defaultArgs, args...)
}

func (kc *Kubectl) Command(args ...string) *exec.Cmd {
	kubectlArgs := kc.Arguments(args...)
	kc.logger.Info("running", "path", kc.kubectlPath, "args", kubectlArgs)
	return exec.Command(kc.kubectlPath, kubectlArgs...)
}

func StartWithStreamOutput(cmd *exec.Cmd) (stdout, stderr io.ReadCloser, err error) {
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return
	}
	stderr, err = cmd.StderrPipe()
	if err != nil {
		return
	}
	err = cmd.Start()
	return
}
