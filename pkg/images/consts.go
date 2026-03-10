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
	SchedulerPluginSchedulerDefaultImageTag  = "registry.k8s.io/scheduler-plugins/kube-scheduler:v0.32.7"
	SchedulerPluginControllerDefaultImageTag = "registry.k8s.io/scheduler-plugins/controller:v0.32.7"
	NodeFeatureDiscoveryDefaultImageTag      = "registry.k8s.io/nfd/node-feature-discovery:v0.15.1"
	ResourceTopologyExporterDefaultImageTag  = "quay.io/k8stopologyawareschedwg/resource-topology-exporter:v0.21.8"
)

const (
	SchedulerPluginSchedulerDefaultImageSHA  = "registry.k8s.io/scheduler-plugins/kube-scheduler@sha256:8d778457bdb7e98366f3b1e70b13c06d7d370564eb94e9c446ba5df7ac73be41"
	SchedulerPluginControllerDefaultImageSHA = "registry.k8s.io/scheduler-plugins/controller@sha256:d4540ec83b110b9e2c77b484a67d18140282581bd761ee25dc00a57d14b9d8ec"
	NodeFeatureDiscoveryDefaultImageSHA      = "registry.k8s.io/nfd/node-feature-discovery@sha256:cab8506a76c96a4318d4cb1858ead6fe55a2e0499f69b4201b01d69d4fa14f10"
	ResourceTopologyExporterDefaultImageSHA  = "quay.io/k8stopologyawareschedwg/resource-topology-exporter@sha256:1e26398bae182f748c44a35fec70b6e7116ff0480c5187748318b9b3abb78919"
)
