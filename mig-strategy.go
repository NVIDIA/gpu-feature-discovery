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
	MigStrategyNone = "none"
)

// MigStrategy defines the strategy to use for setting labels on MIG devices.
type MigStrategy interface {
	GenerateLabels() (map[string]string, error)
}

// NewMigStrategy creates a new MIG strategy to generate labels with.
func NewMigStrategy(strategy string, nvml Nvml) (MigStrategy, error) {
	switch strategy {
	case MigStrategyNone:
		return &migStrategyNone{nvml}, nil
	}
	return nil, fmt.Errorf("Unknown strategy: %v", strategy)
}

type migStrategyNone struct{ nvml Nvml }

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

// getAllMigDevices() across all full GPUs
func getAllMigDevices(nvml Nvml) ([]NvmlDevice, error) {
	n, err := nvml.GetDeviceCount()
	if err != nil {
		return nil, err
	}

	var migs []NvmlDevice
	for i := uint(0); i < n; i++ {
		d, err := nvml.NewDevice(i)
		if err != nil {
			return nil, err
		}

		migEnabled, err := d.IsMigEnabled()
		if err != nil {
			return nil, err
		}

		if !migEnabled {
			continue
		}

		devs, err := d.GetMigDevices()
		if err != nil {
			return nil, err
		}

		migs = append(migs, devs...)
	}

	return migs, nil
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
