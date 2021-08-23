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

package images

import (
	"fmt"
	"os"
	"strings"
)

const (
	Registry     = "quay.io"
	Organization = "k8stopologyawareschedwg"
)

func init() {
	if schedImage, ok := os.LookupEnv("TAS_SCHEDULER_PLUGIN_IMAGE"); ok {
		SchedulerPluginSchedulerImage = schedImage
	}
	if schedCtrlImage, ok := os.LookupEnv("TAS_SCHEDULER_PLUGIN_CONTROLLER_IMAGE"); ok {
		SchedulerPluginControllerImage = schedCtrlImage
	}
	if rteImage, ok := os.LookupEnv("TAS_RESOURCE_EXPORTER_IMAGE"); ok {
		ResourceTopologyExporterImage = rteImage
	}
}

var (
	SchedulerPluginSchedulerImage  = SchedulerPluginSchedulerDefaultImage
	SchedulerPluginControllerImage = SchedulerPluginSchedulerDefaultImage
	ResourceTopologyExporterImage  = ResourceTopologyExporterDefaultImage
)

type Images struct {
	SchedulerPluginScheduler  string
	SchedulerPluginController string
	ResourceTopologyExporter  string
}

func (im Images) ToStrings() []string {
	return []string{
		im.SchedulerPluginController,
		im.SchedulerPluginScheduler,
		im.ResourceTopologyExporter,
	}
}

func Current() Images {
	imgs := Upstream()
	return Images{
		SchedulerPluginScheduler:  Mirror(imgs.SchedulerPluginScheduler),
		SchedulerPluginController: Mirror(imgs.SchedulerPluginController),
		ResourceTopologyExporter:  Mirror(imgs.ResourceTopologyExporter),
	}
}

func Defaults() Images {
	return Images{
		SchedulerPluginScheduler:  SchedulerPluginSchedulerDefaultImage,
		SchedulerPluginController: SchedulerPluginControllerDefaultImage,
		ResourceTopologyExporter:  ResourceTopologyExporterDefaultImage,
	}
}

func Upstream() Images {
	return Images{
		SchedulerPluginScheduler:  SchedulerPluginSchedulerImage,
		SchedulerPluginController: SchedulerPluginControllerImage,
		ResourceTopologyExporter:  ResourceTopologyExporterImage,
	}
}

func Mirror(pullSpec string) string {
	if strings.Contains(pullSpec, "@sha256") {
		// these are safe already, no need to do any logic
		return pullSpec
	}
	f := func(c rune) bool {
		return c == '/' || c == ':'
	}
	components := strings.FieldsFunc(pullSpec, f)
	num := len(components)
	if num < 3 {
		// host, name, tag
		return pullSpec
	}
	tag := fmt.Sprintf("r%d-%s", Revision, components[num-1])
	name := strings.Join(components[1:num-1], "-")
	return fmt.Sprintf("%s/%s/%s:%s", Registry, Organization, name, tag)
}
