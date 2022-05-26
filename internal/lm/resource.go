/**
# Copyright (c) 2022, NVIDIA CORPORATION.  All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
**/

package lm

import (
	"fmt"
	"strings"

	"github.com/NVIDIA/gpu-feature-discovery/internal/nvml"
	spec "github.com/NVIDIA/k8s-device-plugin/api/config/v1"
)

// NewGPUResourceLabeler creates a resource labeler for the specified full GPU device with the specified count
func NewGPUResourceLabeler(config *spec.Config, device nvml.Device, count int) (Labeler, error) {
	if count == 0 {
		return empty{}, nil
	}

	model, err := device.GetName()
	if err != nil {
		return nil, fmt.Errorf("failed to get device model: %v", err)
	}

	memoryInfo, err := device.GetMemoryInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info for device: %v", err)
	}

	fullGPUResourceName, err := spec.NewResourceName("gpu")
	if err != nil {
		return nil, fmt.Errorf("failed to construct resource name: %v", err)
	}

	resourceLabeler := resourceLabeler{
		resourceName: fullGPUResourceName,
		config:       config,
	}

	memoryLabeler := (Labeler)(&empty{})
	if memoryInfo.Total != 0 {
		memoryLabeler = resourceLabeler.single("memory", memoryInfo.Total)
	}

	labelers := Merge(
		newProductLabeler(resourceLabeler, model),
		newCountLabeler(resourceLabeler, count),
		memoryLabeler,
	)

	return labelers, nil
}

// NewMIGResourceLabeler creates a resource labeler for the specified full GPU device with the specified resource name.
func NewMIGResourceLabeler(resourceName spec.ResourceName, config *spec.Config, device nvml.Device, count int) (Labeler, error) {
	if count == 0 {
		return empty{}, nil
	}

	model, err := device.GetName()
	if err != nil {
		return nil, fmt.Errorf("failed to get device model: %v", err)
	}

	migProfile, err := getMigDeviceName(device)
	if err != nil {
		return nil, fmt.Errorf("failed to get MIG profile name: %v", err)
	}

	resourceLabeler := resourceLabeler{
		resourceName: resourceName,
		config:       config,
	}

	attributeLabeler := migAttributeLabeler{
		resourceLabeler: resourceLabeler,
		device:          device,
	}

	labelers := Merge(
		newProductLabeler(resourceLabeler, model, "MIG", migProfile),
		newCountLabeler(resourceLabeler, int(count)),
		attributeLabeler,
	)

	return labelers, nil
}

type resourceLabeler struct {
	resourceName spec.ResourceName
	config       *spec.Config
}

// single creates a single label for the resource. The label key is
// <fully-qualified-resource-name>.suffix
func (rl resourceLabeler) single(suffix string, value interface{}) Labels {
	return rl.labels(map[string]interface{}{suffix: value})

}

// labels creates a set of labels from the specified map for the resource.
// Each key in the map corresponds to a label <fully-qualified-resource-name>.key
func (rl resourceLabeler) labels(suffixValues map[string]interface{}) Labels {
	labels := make(Labels)
	for suffix, value := range suffixValues {
		rl.updateLabel(labels, suffix, value)
	}

	return labels
}

// updateLabel modifies the specified labels, updating <fully-qualified-resource-name>.suffix with
// the provided value.
func (rl resourceLabeler) updateLabel(labels Labels, suffix string, value interface{}) {
	key := rl.key(suffix)

	labels[key] = fmt.Sprintf("%v", value)
}

// key generates the label key for the specified suffix. The key is generated as
// <fully-qualified-resource-name>.suffix
func (rl resourceLabeler) key(suffix string) string {
	return string(rl.resourceName) + "." + suffix
}

// replicationInfo searches the associated config for the resource and returns the replication info
func (rl resourceLabeler) replicationInfo() *spec.ReplicatedResource {
	name := rl.resourceName
	for _, r := range rl.config.Sharing.TimeSlicing.Resources {
		if r.Name == spec.ResourceName(name) {
			return &r
		}
	}
	return nil
}

func newProductLabeler(rl resourceLabeler, parts ...string) Labeler {
	var strippedParts []string
	for _, p := range parts {
		if p != "" {
			strippedParts = append(strippedParts, strings.Replace(p, " ", "-", -1))
		}
	}

	if len(strippedParts) == 0 {
		return empty{}
	}

	if r := rl.replicationInfo(); r != nil && r.Replicas > 1 && r.Rename == "" {
		strippedParts = append(strippedParts, "SHARED")
	}

	return rl.single("product", strings.Join(strippedParts, "-"))
}

func newCountLabeler(rl resourceLabeler, count int) Labeler {
	replicas := 1
	if r := rl.replicationInfo(); r != nil && r.Replicas > 1 {
		replicas = r.Replicas
	}

	labels := make(Labels)
	rl.updateLabel(labels, "count", count*replicas)

	if replicas > 1 {
		rl.updateLabel(labels, "replicas", replicas)
	}

	return labels
}

type migAttributeLabeler struct {
	resourceLabeler
	device nvml.Device
}

func (s migAttributeLabeler) Labels() (Labels, error) {
	attributes, err := s.device.GetAttributes()
	if err != nil {
		return nil, fmt.Errorf("unable to get attributes of MIG device: %v", err)
	}

	labels := s.resourceLabeler.labels(map[string]interface{}{
		"memory":          attributes.MemorySizeMB,
		"multiprocessors": attributes.MultiprocessorCount,
		"slices.gi":       attributes.GpuInstanceSliceCount,
		"slices.ci":       attributes.ComputeInstanceSliceCount,
		"engines.copy":    attributes.SharedCopyEngineCount,
		"engines.decoder": attributes.SharedDecoderCount,
		"engines.encoder": attributes.SharedEncoderCount,
		"engines.jpeg":    attributes.SharedJpegCount,
		"engines.ofa":     attributes.SharedOfaCount,
	})

	return labels, nil
}
