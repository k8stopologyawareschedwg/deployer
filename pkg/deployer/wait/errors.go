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
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func deletionStatusFromError(logger logr.Logger, kind string, key ObjectKey, err error) (bool, error) {
	if err == nil {
		logger.Info("object still present", "kind", kind, "key", key.String())
		return false, nil
	}
	if apierrors.IsNotFound(err) {
		logger.Info("object is gone", "kind", kind, "key", key.String())
		return true, nil
	}
	logger.Info("failed to get object", "kind", kind, "key", key.String(), "error", err)
	return false, err
}
