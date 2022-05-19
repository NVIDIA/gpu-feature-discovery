/**
# Copyright (c) 2019-2022, NVIDIA CORPORATION.  All rights reserved.
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

package nvml

import (
	"fmt"

	"github.com/NVIDIA/go-nvml/pkg/nvml"
)

// NvmlMock : Implementation of Nvml using mocked calls
type NvmlMock struct {
	Devices       []NvmlMockDevice
	DriverVersion string
	CudaMajor     uint
	CudaMinor     uint
	ErrorOnInit   bool
}

// NvmlMockDevice : Implementation of NvmlDevice using mocked calls
type NvmlMockDevice struct {
	Handle       *nvml.Device
	Attributes   *DeviceAttributes
	MigEnabled   bool
	MigDevices   []NvmlMockDevice
	Model        string
	ComputeMajor int
	ComputeMinor int
	TotalMemory  uint64
	UUID         string
}

var _ NvmlDevice = (*NvmlMockDevice)(nil)

// AsInitError creates an NvmlInitError
func (nvmlMock Mock) AsInitError(err error) NvmlInitError {
	return NvmlInitError{err}
}

// IsInitError checks if the specified error is an init error
func (nvmlMock Mock) IsInitError(err error) bool {
	_, isInitError := err.(NvmlInitError)
	return isInitError
}

// Init : Init the mock
func (nvmlMock NvmlMock) Init() error {
	if nvmlMock.ErrorOnInit {
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
	return uint(len(nvmlMock.Devices)), nil
}

// NewDevice : Get information about a fake GPU
func (nvmlMock NvmlMock) NewDevice(id uint) (NvmlDevice, error) {
	if int(id) < len(nvmlMock.Devices) {
		return nvmlMock.Devices[id], nil
	}
	return nil, fmt.Errorf("invalid index: %d", id)
}

// GetDriverVersion : Return a fake driver version
func (nvmlMock NvmlMock) GetDriverVersion() (string, error) {
	return nvmlMock.DriverVersion, nil
}

// GetCudaDriverVersion : Return a fake cuda version
func (nvmlMock NvmlMock) GetCudaDriverVersion() (*uint, *uint, error) {
	return &nvmlMock.CudaMajor, &nvmlMock.CudaMinor, nil
}

// Instance : Return the underlying NVML device instance
func (d NvmlMockDevice) Instance() *nvml.Device {
	return d.Handle
}

// IsMigEnabled : Returns whether MIG is enabled on the device or not
func (d NvmlMockDevice) IsMigEnabled() (bool, error) {
	return d.MigEnabled, nil
}

// GetMigDevices : Returns the list of MIG devices configured on this device
func (d NvmlMockDevice) GetMigDevices() ([]NvmlDevice, error) {
	var devices []NvmlDevice
	for _, m := range d.MigDevices {
		devices = append(devices, m)
	}
	return devices, nil
}

// GetAttributes : Returns the set of of Devices attributes
func (d NvmlMockDevice) GetAttributes() (DeviceAttributes, error) {
	return *d.Attributes, nil
}

// GetCudaComputeCapability returns the mocked CUDA Compute capability
func (d NvmlMockDevice) GetCudaComputeCapability() (int, int, error) {
	return d.ComputeMajor, d.ComputeMinor, nil
}

// GetMemoryInfo returns the mocked memory info
func (d NvmlMockDevice) GetMemoryInfo() (nvml.Memory, error) {
	return nvml.Memory{Total: d.TotalMemory}, nil
}

// GetName returns the mocked device name
func (d NvmlMockDevice) GetName() (string, error) {
	return d.Model, nil
}

// GetUUID returns the mocked device uuid
func (d NvmlMockDevice) GetUUID() (string, error) {
	return d.UUID, nil
}
