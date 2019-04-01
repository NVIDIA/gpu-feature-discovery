// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package main

import "github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"

type NvmlInterface interface {
	Init() error
	Shutdown() error
	GetDeviceCount() (uint, error)
	NewDevice(id uint) (device *nvml.Device, err error)
	GetDriverVersion() (string, error)
}

type NvmlLib struct {
}

func (nvmlLib NvmlLib) Init() error {
	return nvml.Init()
}

func (nvmlLib NvmlLib) Shutdown() error {
	return nvml.Shutdown()
}

func (nvmlLib NvmlLib) GetDeviceCount() (uint, error) {
	return nvml.GetDeviceCount()
}

func (nvmlLib NvmlLib) NewDevice(id uint) (device *nvml.Device, err error) {
	return nvml.NewDevice(id)
}

func (nvmlLib NvmlLib) GetDriverVersion() (string, error) {
	return nvml.GetDriverVersion()
}
