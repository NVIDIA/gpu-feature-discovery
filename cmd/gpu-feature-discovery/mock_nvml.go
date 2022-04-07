// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package main

import (
	"fmt"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

// NvmlMock : Implementation of Nvml using mocked calls
type NvmlMock struct {
	devices       []NvmlMockDevice
	driverVersion string
	cudaMajor     uint
	cudaMinor     uint
	errorOnInit   bool
}

// NvmlMockDevice : Implementation of NvmlDevice using mocked calls
type NvmlMockDevice struct {
	instance     *nvml.Device
	attributes   *nvml.DeviceAttributes
	migEnabled   bool
	migDevices   []NvmlMockDevice
	model        string
	computeMajor int
	computeMinor int
	totalMemory  uint64
	uuid         string
}

var _ NvmlDevice = (*NvmlMockDevice)(nil)

// Init : Init the mock
func (nvmlMock NvmlMock) Init() error {
	if nvmlMock.errorOnInit {
		return fmt.Errorf("nvmlMock error on init")
	}
	return nil
}

// Shutdown : Shutdown the mock
func (nvmlMock NvmlMock) Shutdown() error {
	return nil
}

// GetDeviceCount : Return a fake number of devices
func (nvmlMock NvmlMock) GetDeviceCount() (uint, error) {
	return uint(len(nvmlMock.devices)), nil
}

// NewDevice : Get information about a fake GPU
func (nvmlMock NvmlMock) NewDevice(id uint) (NvmlDevice, error) {
	if int(id) < len(nvmlMock.devices) {
		return nvmlMock.devices[id], nil
	}
	return nil, fmt.Errorf("invalid index: %d", id)
}

// GetDriverVersion : Return a fake driver version
func (nvmlMock NvmlMock) GetDriverVersion() (string, error) {
	return nvmlMock.driverVersion, nil
}

// GetCudaDriverVersion : Return a fake cuda version
func (nvmlMock NvmlMock) GetCudaDriverVersion() (*uint, *uint, error) {
	return &nvmlMock.cudaMajor, &nvmlMock.cudaMinor, nil
}

// Instance : Return the underlying NVML device instance
func (d NvmlMockDevice) Instance() *nvml.Device {
	return d.instance
}

// IsMigEnabled : Returns whether MIG is enabled on the device or not
func (d NvmlMockDevice) IsMigEnabled() (bool, error) {
	return d.migEnabled, nil
}

// GetMigDevices : Returns the list of MIG devices configured on this device
func (d NvmlMockDevice) GetMigDevices() ([]NvmlDevice, error) {
	var devices []NvmlDevice
	for _, m := range d.migDevices {
		devices = append(devices, m)
	}
	return devices, nil
}

// GetAttributes : Returns the set of of Devices attributes
func (d NvmlMockDevice) GetAttributes() (nvml.DeviceAttributes, error) {
	return *d.attributes, nil
}

// GetCudaComputeCapability returns the mocked CUDA Compute capability
func (d NvmlMockDevice) GetCudaComputeCapability() (int, int, error) {
	return d.computeMajor, d.computeMinor, nil
}

// GetMemoryInfo returns the mocked memory info
func (d NvmlMockDevice) GetMemoryInfo() (nvml.Memory, error) {
	return nvml.Memory{Total: d.totalMemory}, nil
}

// GetName returns the mocked device name
func (d NvmlMockDevice) GetName() (string, error) {
	return d.model, nil
}

// GetUUID returns the mocked device uuid
func (d NvmlMockDevice) GetUUID() (string, error) {
	return d.uuid, nil
}
