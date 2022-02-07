// Copyright (c) 2021, NVIDIA CORPORATION. All rights reserved.

package main

import (
	"fmt"
)

// MIGCapableDevices stores information about all devices on the node
type MIGCapableDevices struct {
	// The NVML library
	nvml Nvml
	// devicesMap holds a list of devices, separated by whether they have MigEnabled or not
	devicesMap map[bool][]NvmlDevice
}

// NewMIGCapableDevices creates a new MIGCapableDevices struct and returns a pointer to it.
func NewMIGCapableDevices(nvml Nvml) *MIGCapableDevices {
	return &MIGCapableDevices{
		nvml:       nvml,
		devicesMap: nil, // Is initialized on first use
	}
}

func (devices *MIGCapableDevices) getDevicesMap() (map[bool][]NvmlDevice, error) {
	if devices.devicesMap == nil {
		n, err := devices.nvml.GetDeviceCount()
		if err != nil {
			return nil, err
		}

		migEnabledDevicesMap := make(map[bool][]NvmlDevice)
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
func (devices *MIGCapableDevices) GetDevicesWithMigEnabled() ([]NvmlDevice, error) {
	devicesMap, err := devices.getDevicesMap()
	if err != nil {
		return nil, err
	}
	return devicesMap[true], nil
}

// GetDevicesWithMigDisabled returns a list of devices with migEnabled=false
func (devices *MIGCapableDevices) GetDevicesWithMigDisabled() ([]NvmlDevice, error) {
	devicesMap, err := devices.getDevicesMap()
	if err != nil {
		return nil, err
	}
	return devicesMap[false], nil
}

// AssertAllMigEnabledDevicesAreValid ensures that all devices with migEnabled=true are valid. This means:
// * The have at least 1 mig devices associated with them
// Returns nill if the device is valid, or an error if these are not valid
func (devices *MIGCapableDevices) AssertAllMigEnabledDevicesAreValid() error {
	devicesMap, err := devices.getDevicesMap()
	if err != nil {
		return err
	}

	for _, d := range devicesMap[true] {
		migs, err := d.GetMigDevices()
		if err != nil {
			return err
		}
		if len(migs) == 0 {
			return fmt.Errorf("No MIG devices associated with %v: %v", d.Instance().Path, d.Instance().UUID)
		}
	}
	return nil
}

// GetAllMigDevices returns a list of all MIG devices.
func (devices *MIGCapableDevices) GetAllMigDevices() ([]NvmlDevice, error) {
	devicesMap, err := devices.getDevicesMap()
	if err != nil {
		return nil, err
	}

	var migs []NvmlDevice
	for _, d := range devicesMap[true] {
		devs, err := d.GetMigDevices()
		if err != nil {
			return nil, err
		}
		migs = append(migs, devs...)
	}
	return migs, nil
}
