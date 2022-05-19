/**
# Copyright (c) 2021-2022, NVIDIA CORPORATION.  All rights reserved.
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

package main

import "github.com/NVIDIA/gpu-feature-discovery/internal/nvml"

// MIGCapableDevices stores information about all devices on the node
type MIGCapableDevices struct {
	// The NVML library
	nvml nvml.Nvml
	// devicesMap holds a list of devices, separated by whether they have MigEnabled or not
	devicesMap map[bool][]nvml.NvmlDevice
}

// NewMIGCapableDevices creates a new MIGCapableDevices struct and returns a pointer to it.
func NewMIGCapableDevices(nvml nvml.Nvml) *MIGCapableDevices {
	return &MIGCapableDevices{
		nvml:       nvml,
		devicesMap: nil, // Is initialized on first use
	}
}

func (devices *MIGCapableDevices) getDevicesMap() (map[bool][]nvml.NvmlDevice, error) {
	if devices.devicesMap == nil {
		n, err := devices.nvml.GetDeviceCount()
		if err != nil {
			return nil, err
		}

		migEnabledDevicesMap := make(map[bool][]nvml.NvmlDevice)
		for i := uint(0); i < n; i++ {
			d, err := devices.nvml.NewDevice(i)
			if err != nil {
				return nil, err
			}

			isMigEnabled, err := d.IsMigEnabled()
			if err != nil {
				return nil, err
			}

			migEnabledDevicesMap[isMigEnabled] = append(migEnabledDevicesMap[isMigEnabled], d)
		}

		devices.devicesMap = migEnabledDevicesMap
	}
	return devices.devicesMap, nil
}

// GetDevicesWithMigEnabled returns a list of devices with migEnabled=true
func (devices *MIGCapableDevices) GetDevicesWithMigEnabled() ([]nvml.NvmlDevice, error) {
	devicesMap, err := devices.getDevicesMap()
	if err != nil {
		return nil, err
	}
	return devicesMap[true], nil
}

// GetDevicesWithMigDisabled returns a list of devices with migEnabled=false
func (devices *MIGCapableDevices) GetDevicesWithMigDisabled() ([]nvml.NvmlDevice, error) {
	devicesMap, err := devices.getDevicesMap()
	if err != nil {
		return nil, err
	}
	return devicesMap[false], nil
}

// AnyMigEnabledDeviceIsEmpty checks whether at least one MIG device has no MIG devices configured
func (devices *MIGCapableDevices) AnyMigEnabledDeviceIsEmpty() (bool, error) {
	devicesMap, err := devices.getDevicesMap()
	if err != nil {
		return false, err
	}

	if len(devicesMap[true]) == 0 {
		// By definition the property is true for the empty set
		return true, nil
	}

	for _, d := range devicesMap[true] {
		migs, err := d.GetMigDevices()
		if err != nil {
			return false, err
		}
		if len(migs) == 0 {
			return true, nil
		}
	}
	return false, nil
}

// GetAllMigDevices returns a list of all MIG devices.
func (devices *MIGCapableDevices) GetAllMigDevices() ([]nvml.NvmlDevice, error) {
	devicesMap, err := devices.getDevicesMap()
	if err != nil {
		return nil, err
	}

	var migs []nvml.NvmlDevice
	for _, d := range devicesMap[true] {
		devs, err := d.GetMigDevices()
		if err != nil {
			return nil, err
		}
		migs = append(migs, devs...)
	}
	return migs, nil
}
