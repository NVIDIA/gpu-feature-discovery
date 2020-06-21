// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package main

import "github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"

// Nvml : Type to represent interactions with NVML
type Nvml interface {
	Init() error
	Shutdown() error
	GetDeviceCount() (uint, error)
	NewDevice(id uint) (device *nvml.Device, err error)
	GetDriverVersion() (string, error)
	GetCudaDriverVersion() (*uint, *uint, error)
}

// NvmlLib : Implementation of Nvml using the NVML lib
type NvmlLib struct{}

// Init : Init NVML lib
func (nvmlLib NvmlLib) Init() error {
	return nvml.Init()
}

// Shutdown : Shutdown NVML lib
func (nvmlLib NvmlLib) Shutdown() error {
	return nvml.Shutdown()
}

// GetDeviceCount : Return the number of GPUs using NVML
func (nvmlLib NvmlLib) GetDeviceCount() (uint, error) {
	return nvml.GetDeviceCount()
}

// NewDevice : Get all information about a GPU using NVML
func (nvmlLib NvmlLib) NewDevice(id uint) (device *nvml.Device, err error) {
	return nvml.NewDevice(id)
}

// GetDriverVersion : Return the driver version using NVML
func (nvmlLib NvmlLib) GetDriverVersion() (string, error) {
	return nvml.GetDriverVersion()
}

// GetCudaDriverVersion : Return the cuda version using NVML
func (nvmlLib NvmlLib) GetCudaDriverVersion() (*uint, *uint, error) {
	return nvml.GetCudaDriverVersion()
}
