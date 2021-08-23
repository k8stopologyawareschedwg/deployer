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

package main

import (
	"fmt"
	"os"

	"github.com/k8stopologyawareschedwg/deployer/pkg/images"
)

func main() {
	runtime := "podman"
	if val, ok := os.LookupEnv("CONTAINER_RUNTUME"); ok {
		runtime = val
	}
	imgs := images.Upstream().ToStrings()

	fmt.Printf("#!/bin/bash\n")
	fmt.Printf("set -uex\n")

	for _, pullSpec := range imgs {
		fmt.Printf("%s pull %s\n", runtime, pullSpec)
	}

	for _, pullSpec := range imgs {
		fmt.Printf("%s tag %s %s\n", runtime, pullSpec, images.Mirror(pullSpec))
	}

	for _, pullSpec := range imgs {
		fmt.Printf("%s push %s\n", runtime, images.Mirror(pullSpec))
	}
}
