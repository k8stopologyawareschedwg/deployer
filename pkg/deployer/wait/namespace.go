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
 * Copyright 2023 Red Hat, Inc.
 */

package wait

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	k8swait "k8s.io/apimachinery/pkg/util/wait"
)

func (wt Waiter) ForNamespaceDeleted(ctx context.Context, namespace string) error {
	log := wt.Log.WithValues("namespace", namespace)
	log.Info("wait for the namespace to be gone")
	return k8swait.PollImmediate(wt.PollInterval, wt.PollTimeout, func() (bool, error) {
		nsKey := ObjectKey{Name: namespace}
		ns := corev1.Namespace{} // unused
		err := wt.Cli.Get(ctx, nsKey.AsKey(), &ns)
		return deletionStatusFromError(wt.Log, "Namespace", nsKey, err)
	})
}
