// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package main

import "github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"

// NvmlMock : Implementation of NvmlInterface using mocked calls
type NvmlMock struct {
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
	return 1, nil
}

// NewDevice : Get information about a fake GPU
func (nvmlMock NvmlMock) NewDevice(id uint) (*nvml.Device, error) {
	device := nvml.Device{}
	one := 1
	model := "MOCKMODEL"
	memory := uint64(128)
	device.Model = &model
	device.Memory = &memory
	device.CudaComputeCapability.Major = &one
	device.CudaComputeCapability.Minor = &one
	return &device, nil
}

// GetDriverVersion : Return a fake driver version
func (nvmlMock NvmlMock) GetDriverVersion() (string, error) {
	return "400.300", nil
}

// GetCudaDriverVersion : Return a fake cuda version
func (nvmlMock NvmlMock) GetCudaDriverVersion() (*uint, *uint, error) {
	major := uint(1)
	minor := uint(1)
	return &major, &minor, nil
}
