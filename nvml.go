// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package main

import "github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"

// Nvml : Type to represent interactions with NVML
type Nvml interface {
	Init() error
	Shutdown() error
	GetDeviceCount() (uint, error)
	NewDevice(id uint) (device NvmlDevice, err error)
	GetDriverVersion() (string, error)
	GetCudaDriverVersion() (*uint, *uint, error)
}

// NvmlDevice : Type to represent interactions with an nvml.Device
type NvmlDevice interface {
	Instance() *nvml.Device
	IsMigEnabled() (bool, error)
	GetMigDevices() ([]NvmlDevice, error)
	GetAttributes() (nvml.DeviceAttributes, error)
}

// NvmlLib : Implementation of Nvml using the NVML lib
type NvmlLib struct{}

// NvmlLibDevice : Implementation of NvmlDevice using a device from the NVML lib
type NvmlLibDevice struct {
	device *nvml.Device
}

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
func (nvmlLib NvmlLib) NewDevice(id uint) (device NvmlDevice, err error) {
	d, err := nvml.NewDevice(id)
	if err != nil {
		return nil, err
	}
	return NvmlLibDevice{d}, err
}

// GetDriverVersion : Return the driver version using NVML
func (nvmlLib NvmlLib) GetDriverVersion() (string, error) {
	return nvml.GetDriverVersion()
}

// GetCudaDriverVersion : Return the cuda version using NVML
func (nvmlLib NvmlLib) GetCudaDriverVersion() (*uint, *uint, error) {
	return nvml.GetCudaDriverVersion()
}

// Instance : Return the underlying NVML device instance
func (d NvmlLibDevice) Instance() *nvml.Device {
	return d.device
}

// IsMigEnabled : Returns whether MIG is enabled on the device or not
func (d NvmlLibDevice) IsMigEnabled() (bool, error) {
	return d.device.IsMigEnabled()
}

// GetMigDevices : Returns the list of MIG devices configured on this device
func (d NvmlLibDevice) GetMigDevices() ([]NvmlDevice, error) {
	devs, err := d.device.GetMigDevices()
	if err != nil {
		return nil, err
	}

	var migs []NvmlDevice
	for _, d := range devs {
		migs = append(migs, NvmlLibDevice{d})
	}

	return migs, nil
}

// GetAttributes : Returns the set of of Devices attributes
func (d NvmlLibDevice) GetAttributes() (nvml.DeviceAttributes, error) {
	return d.device.GetAttributes()
}
