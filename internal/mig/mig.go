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

package mig

import (
	"github.com/NVIDIA/gpu-feature-discovery/internal/resource"
)

// DeviceInfo stores information about all devices on the node
type DeviceInfo struct {
	// The NVML library
	manager resource.Manager
	// devicesMap holds a list of devices, separated by whether they have MigEnabled or not
	devicesMap map[bool][]resource.Device
}

// NewDeviceInfo creates a new DeviceInfo struct and returns a pointer to it.
func NewDeviceInfo(manager resource.Manager) *DeviceInfo {
	return &DeviceInfo{
		manager:    manager,
		devicesMap: nil, // Is initialized on first use
	}
}

// GetDevicesMap returns the list of devices separated by whether they have MIG enabled.
// The first call will construct the map.
func (devices *DeviceInfo) GetDevicesMap() (map[bool][]resource.Device, error) {
	if devices.devicesMap == nil {
		n, err := devices.manager.GetDeviceCount()
		if err != nil {
			return nil, err
		}

		migEnabledDevicesMap := make(map[bool][]resource.Device)
		for i := 0; i < n; i++ {
			d, err := devices.manager.GetDeviceByIndex(i)
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
func (devices *DeviceInfo) GetDevicesWithMigEnabled() ([]resource.Device, error) {
	devicesMap, err := devices.GetDevicesMap()
	if err != nil {
		return nil, err
	}
	return devicesMap[true], nil
}

// GetDevicesWithMigDisabled returns a list of devices with migEnabled=false
func (devices *DeviceInfo) GetDevicesWithMigDisabled() ([]resource.Device, error) {
	devicesMap, err := devices.GetDevicesMap()
	if err != nil {
		return nil, err
	}
	return devicesMap[false], nil
}

// AnyMigEnabledDeviceIsEmpty checks whether at least one MIG device has no MIG devices configured
func (devices *DeviceInfo) AnyMigEnabledDeviceIsEmpty() (bool, error) {
	devicesMap, err := devices.GetDevicesMap()
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
func (devices *DeviceInfo) GetAllMigDevices() ([]resource.Device, error) {
	devicesMap, err := devices.GetDevicesMap()
	if err != nil {
		return nil, err
	}

	var migs []resource.Device
	for _, d := range devicesMap[true] {
		devs, err := d.GetMigDevices()
		if err != nil {
			return nil, err
		}
		migs = append(migs, devs...)
	}
	return migs, nil
}
