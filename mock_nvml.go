// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package main

import (
	"fmt"
	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
)

// NvmlMock : Implementation of Nvml using mocked calls
type NvmlMock struct {
	devices       []NvmlMockDevice
	driverVersion string
	cudaMajor     uint
	cudaMinor     uint
}

// NvmlMockDevice : Implementation of NvmlDevice using mocked calls
type NvmlMockDevice struct {
	instance *nvml.Device
}

// Init : Init the mock
func (nvmlMock NvmlMock) Init() error {
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
	if int(id) <= len(nvmlMock.devices) {
		return nvmlMock.devices[id], nil
	}
	return nil, fmt.Errorf("Invalid index: %d", id)
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
