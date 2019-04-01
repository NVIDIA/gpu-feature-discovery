// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package main

import "github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"

type NvmlMock struct {
}

func (nvmlMock NvmlMock) Init() error {
	return nil
}

func (nvmlMock NvmlMock) Shutdown() error {
	return nil
}

func (nvmlMock NvmlMock) GetDeviceCount() (uint, error) {
	return 1, nil
}

func (nvmlMock NvmlMock) NewDevice(id uint) (*nvml.Device, error) {
	device := nvml.Device{}
	model := "MOCK-MODEL"
	memory := uint64(128)
	device.Model = &model
	device.Memory = &memory
	return &device, nil
}

func (nvmlMock NvmlMock) GetDriverVersion() (string, error) {
	return "MOCK-DRIVER-VERSION", nil
}
