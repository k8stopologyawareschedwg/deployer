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

package wait

import (
	"fmt"
	"time"

	"github.com/go-logr/logr"

	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	DefaultPollInterval = 1 * time.Second
	// DefaultPollTimeout was computed by trial and error, not scientifically,
	// so it may adjusted in the future any time.
	// Roughly match the time it takes for pods to go running in CI.
	DefaultPollTimeout = 3 * time.Minute
)

var (
	basePollInterval = DefaultPollInterval
	basePollTimeout  = DefaultPollTimeout
)

func SetBaseValues(interval, timeout time.Duration) {
	basePollInterval = interval
	basePollTimeout = timeout
}

type ObjectKey struct {
	Namespace string
	Name      string
}

func ObjectKeyFromObject(obj metav1.Object) ObjectKey {
	return ObjectKey{Namespace: obj.GetNamespace(), Name: obj.GetName()}
}

func (ok ObjectKey) AsKey() types.NamespacedName {
	return types.NamespacedName{
		Namespace: ok.Namespace,
		Name:      ok.Name,
	}
}

func (ok ObjectKey) String() string {
	return fmt.Sprintf("%s/%s", ok.Namespace, ok.Name)
}

type Waiter struct {
	Cli          client.Client
	Log          logr.Logger
	PollTimeout  time.Duration
	PollInterval time.Duration
}

func With(cli client.Client, log logr.Logger) *Waiter {
	return &Waiter{
		Cli:          cli,
		Log:          log,
		PollTimeout:  basePollTimeout,
		PollInterval: basePollInterval,
	}
}

func (wt *Waiter) String() string {
	return fmt.Sprintf("wait every %v up to %v", wt.PollInterval, wt.PollTimeout)
}

func (wt *Waiter) Timeout(tt time.Duration) *Waiter {
	wt.PollTimeout = tt
	return wt
}

func (wt *Waiter) Interval(iv time.Duration) *Waiter {
	wt.PollInterval = iv
	return wt
}
