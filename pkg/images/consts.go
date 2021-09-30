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

const (
	SchedulerPluginSchedulerDefaultImage  = "k8s.gcr.io/scheduler-plugins/kube-scheduler:v0.19.9"
	SchedulerPluginControllerDefaultImage = "k8s.gcr.io/scheduler-plugins/controller:v0.19.9"
	ResourceTopologyExporterDefaultImage  = "quay.io/openshift-kni/resource-topology-exporter:v0.2.1"
)
