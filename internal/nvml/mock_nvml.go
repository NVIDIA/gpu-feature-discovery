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

// Mock : Implementation of Nvml using mocked calls
type Mock struct {
	Devices       []MockDevice
	DriverVersion string
	CudaMajor     uint
	CudaMinor     uint
	ErrorOnInit   bool
}

// MockDevice : Implementation of Device using mocked calls
type MockDevice struct {
	Handle       *nvml.Device
	Attributes   *DeviceAttributes
	MigEnabled   bool
	MigDevices   []MockDevice
	Model        string
	ComputeMajor int
	ComputeMinor int
	TotalMemory  uint64
	UUID         string
}

var _ Device = (*MockDevice)(nil)

// Init : Init the mock
func (nvmlMock Mock) Init() error {
	if nvmlMock.ErrorOnInit {
		return fmt.Errorf("nvmlMock error on init")
	}
	return nil
}

// Shutdown : Shutdown the mock
func (nvmlMock Mock) Shutdown() error {
	return nil
}

// GetDeviceCount : Return a fake number of devices
func (nvmlMock Mock) GetDeviceCount() (uint, error) {
	return uint(len(nvmlMock.Devices)), nil
}

// NewDevice : Get information about a fake GPU
func (nvmlMock Mock) NewDevice(id uint) (Device, error) {
	if int(id) < len(nvmlMock.Devices) {
		return nvmlMock.Devices[id], nil
	}
	return nil, fmt.Errorf("invalid index: %d", id)
}

// GetDriverVersion : Return a fake driver version
func (nvmlMock Mock) GetDriverVersion() (string, error) {
	return nvmlMock.DriverVersion, nil
}

// GetCudaDriverVersion : Return a fake cuda version
func (nvmlMock Mock) GetCudaDriverVersion() (*uint, *uint, error) {
	return &nvmlMock.CudaMajor, &nvmlMock.CudaMinor, nil
}

// Instance : Return the underlying NVML device instance
func (d MockDevice) Instance() *nvml.Device {
	return d.Handle
}

// IsMigEnabled : Returns whether MIG is enabled on the device or not
func (d MockDevice) IsMigEnabled() (bool, error) {
	return d.MigEnabled, nil
}

// GetMigDevices : Returns the list of MIG devices configured on this device
func (d MockDevice) GetMigDevices() ([]Device, error) {
	var devices []Device
	for _, m := range d.MigDevices {
		devices = append(devices, m)
	}
	return devices, nil
}

// GetAttributes : Returns the set of of Devices attributes
func (d MockDevice) GetAttributes() (DeviceAttributes, error) {
	return *d.Attributes, nil
}

// GetCudaComputeCapability returns the mocked CUDA Compute capability
func (d MockDevice) GetCudaComputeCapability() (int, int, error) {
	return d.ComputeMajor, d.ComputeMinor, nil
}

// GetMemoryInfo returns the mocked memory info
func (d MockDevice) GetMemoryInfo() (Memory, error) {
	return Memory{Total: d.TotalMemory}, nil
}

// GetName returns the mocked device name
func (d MockDevice) GetName() (string, error) {
	return d.Model, nil
}

// GetUUID returns the mocked device uuid
func (d MockDevice) GetUUID() (string, error) {
	return d.UUID, nil
}

// GetDeviceHandleFromMigDeviceHandle returns the device handle of the parent device
func (d MockDevice) GetDeviceHandleFromMigDeviceHandle() (Device, error) {
	return d, nil
}
