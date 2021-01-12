/*
 * Copyright (c) 2020, NVIDIA CORPORATION.  All rights reserved.
 *
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
 */

package main

import (
	"fmt"
	"strings"
)

// Constants representing different MIG strategies.
const (
	MigStrategyNone   = "none"
	MigStrategySingle = "single"
	MigStrategyMixed  = "mixed"
)

// MigStrategy defines the strategy to use for setting labels on MIG devices.
type MigStrategy interface {
	GenerateLabels() (map[string]string, error)
}

// MigDeviceCounts maintains a count of unique MIG device types across all GPUs on a node
type MigDeviceCounts map[string]int

// NewMigStrategy creates a new MIG strategy to generate labels with.
func NewMigStrategy(strategy string, nvml Nvml) (MigStrategy, error) {
	switch strategy {
	case MigStrategyNone:
		return &migStrategyNone{nvml}, nil
	case MigStrategySingle:
		return &migStrategySingle{nvml}, nil
	case MigStrategyMixed:
		return &migStrategyMixed{nvml}, nil
	}
	return nil, fmt.Errorf("Unknown strategy: %v", strategy)
}

type migStrategyNone struct{ nvml Nvml }
type migStrategySingle struct{ nvml Nvml }
type migStrategyMixed struct{ nvml Nvml }

// migStrategyNone
func (s *migStrategyNone) GenerateLabels() (map[string]string, error) {
	count, err := s.nvml.GetDeviceCount()
	if err != nil {
		return nil, fmt.Errorf("Error getting device count: %v", err)
	}

	device, err := s.nvml.NewDevice(0)
	if err != nil {
		return nil, fmt.Errorf("Error getting device: %v", err)
	}

	labels := make(map[string]string)
	labels["nvidia.com/gpu.count"] = fmt.Sprintf("%d", count)
	if device.Instance().Model != nil {
		model := strings.Replace(*device.Instance().Model, " ", "-", -1)
		labels["nvidia.com/gpu.product"] = model
	}
	if device.Instance().Memory != nil {
		memory := *device.Instance().Memory
		labels["nvidia.com/gpu.memory"] = fmt.Sprintf("%d", memory)
	}

	return labels, nil
}

// migStrategySingle
func (s *migStrategySingle) GenerateLabels() (map[string]string, error) {
	// Generate the same "base" labels as the none strategy
	none, _ := NewMigStrategy(MigStrategyNone, s.nvml)
	labels, err := none.GenerateLabels()
	if err != nil {
		return nil, fmt.Errorf("Unable to generate base labels: %v", err)
	}

	// Add a new label specifying the MIG strategy
	labels["nvidia.com/mig.strategy"] = "single"

	devices := NewMIGCapableDevices(s.nvml)

	migEnabledDevices, err := devices.GetDevicesWithMigEnabled()
	if err != nil {
		return nil, fmt.Errorf("Unabled to retrieve list of MIG-enabled devices: %v", err)
	}
	// No devices have migEnabled=true. This is equivalent to the `none` MIG strategy
	if len(migEnabledDevices) == 0 {
		return labels, nil
	}

	migDisabledDevices, err := devices.GetDevicesWithMigDisabled()
	if err != nil {
		return nil, fmt.Errorf("Unabled to retrieve list of non-MIG-enabled devices: %v", err)
	}
	if len(migDisabledDevices) != 0 {
		return nil, fmt.Errorf("For mig.strategy=single all devices on the node must all be configured with the same migEnabled value")
	}
	if err := devices.AssertAllMigEnabledDevicesAreValid(); err != nil {
		return nil, fmt.Errorf("At least one device with migEnabled=true was not configured corectly: %v", err)
	}

	// Verify that all MIG devices on this node are the same type
	name := ""
	counts := make(MigDeviceCounts)

	migs, err := devices.GetAllMigDevices()
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve list of MIG devices: %v", err)
	}
	for _, mig := range migs {
		name, err = getMigDeviceName(mig)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse MIG device name: %v", err)
		}
		counts[name]++
	}

	if len(counts) == 0 {
		return nil, fmt.Errorf("No MIG devices present on node")
	}

	if len(counts) != 1 {
		return nil, fmt.Errorf("More than one MIG device type present on node")
	}

	// Get the attributes of only the first MIG device (since they are all the same)
	attributes, err := migs[0].GetAttributes()
	if err != nil {
		return nil, fmt.Errorf("Unable to get attributes of MIG device: %v", err)
	}

	// Override some top-level GPU labels set by the 'none' strategy with MIG specific values
	labels["nvidia.com/gpu.count"] = fmt.Sprintf("%d", counts[name])
	labels["nvidia.com/gpu.product"] = fmt.Sprintf("%s-MIG-%s", labels["nvidia.com/gpu.product"], name)
	labels["nvidia.com/gpu.memory"] = fmt.Sprintf("%d", attributes.MemorySizeMB)

	// Add new MIG specific labels on the top-level GPU type
	labels["nvidia.com/gpu.multiprocessors"] = fmt.Sprintf("%d", attributes.MultiprocessorCount)
	labels["nvidia.com/gpu.slices.gi"] = fmt.Sprintf("%d", attributes.GpuInstanceSliceCount)
	labels["nvidia.com/gpu.slices.ci"] = fmt.Sprintf("%d", attributes.ComputeInstanceSliceCount)
	labels["nvidia.com/gpu.engines.copy"] = fmt.Sprintf("%d", attributes.SharedCopyEngineCount)
	labels["nvidia.com/gpu.engines.decoder"] = fmt.Sprintf("%d", attributes.SharedDecoderCount)
	labels["nvidia.com/gpu.engines.encoder"] = fmt.Sprintf("%d", attributes.SharedEncoderCount)
	labels["nvidia.com/gpu.engines.jpeg"] = fmt.Sprintf("%d", attributes.SharedJpegCount)
	labels["nvidia.com/gpu.engines.ofa"] = fmt.Sprintf("%d", attributes.SharedOfaCount)

	return labels, nil
}

// migStrategyMixed
func (s *migStrategyMixed) GenerateLabels() (map[string]string, error) {
	// Generate the same "base" labels as the none strategy
	none, _ := NewMigStrategy(MigStrategyNone, s.nvml)
	labels, err := none.GenerateLabels()
	if err != nil {
		return nil, fmt.Errorf("Unable to generate base labels: %v", err)
	}

	// Add a new label specifying the MIG strategy
	labels["nvidia.com/mig.strategy"] = "mixed"

	devices := NewMIGCapableDevices(s.nvml)

	// Enumerate the MIG devices on this node. In mig.strategy=mixed we ignore devices
	// configured with migEnabled=true but exposing no MIG devices.
	migs, err := devices.GetAllMigDevices()
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve list of MIG devices: %v", err)
	}

	// Add new MIG related labels on each individual MIG type
	counts := make(MigDeviceCounts)
	for _, mig := range migs {
		name, err := getMigDeviceName(mig)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse MIG device name: %v", err)
		}

		// Only set labels for a MIG device type the first time we encounter it
		if counts[name] == 0 {
			attributes, err := mig.GetAttributes()
			if err != nil {
				return nil, fmt.Errorf("Unable to get attributes of MIG device: %v", err)
			}

			prefix := fmt.Sprintf("nvidia.com/mig-%s", name)
			labels[prefix+".memory"] = fmt.Sprintf("%d", attributes.MemorySizeMB)
			labels[prefix+".multiprocessors"] = fmt.Sprintf("%d", attributes.MultiprocessorCount)
			labels[prefix+".slices.gi"] = fmt.Sprintf("%d", attributes.GpuInstanceSliceCount)
			labels[prefix+".slices.ci"] = fmt.Sprintf("%d", attributes.ComputeInstanceSliceCount)
			labels[prefix+".engines.copy"] = fmt.Sprintf("%d", attributes.SharedCopyEngineCount)
			labels[prefix+".engines.decoder"] = fmt.Sprintf("%d", attributes.SharedDecoderCount)
			labels[prefix+".engines.encoder"] = fmt.Sprintf("%d", attributes.SharedEncoderCount)
			labels[prefix+".engines.jpeg"] = fmt.Sprintf("%d", attributes.SharedJpegCount)
			labels[prefix+".engines.ofa"] = fmt.Sprintf("%d", attributes.SharedOfaCount)
		}

		// Maintain the total count of this MIG device type for setting later
		counts[name]++
	}

	// Set the total count on each new MIG type
	for name, count := range counts {
		prefix := fmt.Sprintf("nvidia.com/mig-%s", name)
		labels[prefix+".count"] = fmt.Sprintf("%d", count)
	}

	return labels, nil
}

// getMigDeviceName() returns the canonical name of the MIG device
func getMigDeviceName(mig NvmlDevice) (string, error) {
	attr, err := mig.GetAttributes()
	if err != nil {
		return "", err
	}

	g := attr.GpuInstanceSliceCount
	gb := ((attr.MemorySizeMB + 1000 - 1) / 1000)
	r := fmt.Sprintf("%dg.%dgb", g, gb)

	return r, nil
}
