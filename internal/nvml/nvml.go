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

// Nvml : Type to represent interactions with NVML
type Nvml interface {
	Init() error
	Shutdown() error
	GetDeviceCount() (uint, error)
	NewDevice(id uint) (device Device, err error)
	GetDriverVersion() (string, error)
	GetCudaDriverVersion() (*uint, *uint, error)
}

// Device : Type to represent interactions with an nvml.Device
type Device interface {
	IsMigEnabled() (bool, error)
	GetMigDevices() ([]Device, error)
	GetAttributes() (DeviceAttributes, error)
	GetCudaComputeCapability() (int, int, error)
	GetUUID() (string, error)
	GetName() (string, error)
	GetMemoryInfo() (Memory, error)
	// TODO: This can be cleaned up
	GetArchFamily() (string, error)
}

// Lib : Implementation of Nvml using the NVML lib
type Lib struct{}

// LibDevice : Implementation of Device using a device from the NVML lib
type LibDevice struct {
	device      *nvml.Device
	isMigDevice bool
}

// DeviceAttributes mirrors the nvml Device attributes
type DeviceAttributes nvml.DeviceAttributes

// Memory mirrors the nvml Memory attributes
type Memory nvml.Memory

// Init : Init NVML lib
func (nvmlLib Lib) Init() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("unexpected failure calling nvml.Init: %v", r)
		}
	}()

	ret := nvml.Init()
	if ret != nvml.SUCCESS {
		return errorString(ret)
	}

	return nil
}

// Shutdown : Shutdown NVML lib
func (nvmlLib Lib) Shutdown() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("unexpected failure calling nvml.Shutdown: %v", r)
		}
	}()

	ret := nvml.Shutdown()
	if ret != nvml.SUCCESS {
		return fmt.Errorf("NVML error: %v", nvml.ErrorString(ret))
	}

	return
}

// GetDeviceCount : Return the number of GPUs using NVML
func (nvmlLib Lib) GetDeviceCount() (uint, error) {
	count, ret := nvml.DeviceGetCount()
	if ret != nvml.SUCCESS {
		return 0, errorString(ret)
	}
	return uint(count), nil
}

// NewDevice : Get all information about a GPU using NVML
func (nvmlLib Lib) NewDevice(id uint) (Device, error) {
	h, ret := nvml.DeviceGetHandleByIndex(int(id))
	if ret != nvml.SUCCESS {
		return nil, errorString(ret)
	}

	d := LibDevice{
		device:      &h,
		isMigDevice: false,
	}
	return d, nil
}

// GetDriverVersion : Return the driver version using NVML
func (nvmlLib Lib) GetDriverVersion() (string, error) {
	v, ret := nvml.SystemGetDriverVersion()
	if ret != nvml.SUCCESS {
		return "", errorString(ret)
	}

	return v, nil
}

// GetCudaDriverVersion : Return the cuda v using NVML
func (nvmlLib Lib) GetCudaDriverVersion() (*uint, *uint, error) {
	v, ret := nvml.SystemGetCudaDriverVersion()
	if ret != nvml.SUCCESS {
		return nil, nil, errorString(ret)
	}

	major := uint(v / 1000)
	minor := uint(v % 1000 / 10)

	return &major, &minor, nil
}

// IsMigEnabled : Returns whether MIG is enabled on the device or not.
// Only the current mode is considered and the pending mode is ignored.
func (d LibDevice) IsMigEnabled() (bool, error) {
	cm, _, ret := d.device.GetMigMode()
	if ret == nvml.ERROR_NOT_SUPPORTED {
		return false, nil
	}
	if ret != nvml.SUCCESS {
		return false, errorString(ret)
	}

	return (cm == nvml.DEVICE_MIG_ENABLE), nil
}

// GetMigDevices : Returns the list of MIG devices configured on this device
func (d LibDevice) GetMigDevices() ([]Device, error) {
	n, ret := d.device.GetMaxMigDeviceCount()
	if ret != nvml.SUCCESS {
		return nil, errorString(ret)
	}

	var migs []Device
	for i := 0; i < n; i++ {
		mig, ret := d.device.GetMigDeviceHandleByIndex(i)
		if ret != nvml.ERROR_NOT_FOUND {
			continue
		}
		if ret != nvml.SUCCESS {
			return nil, errorString(ret)
		}

		d := LibDevice{
			device:      &mig,
			isMigDevice: true,
		}
		migs = append(migs, d)
	}
	return migs, nil
}

// GetAttributes : Returns the set of of Devices attributes
func (d LibDevice) GetAttributes() (DeviceAttributes, error) {
	attributes, ret := d.device.GetAttributes()
	if ret != nvml.SUCCESS {
		return DeviceAttributes{}, errorString(ret)
	}

	return DeviceAttributes(attributes), nil
}

// GetCudaComputeCapability returns the CUDA Compute Capability major and minor versions.
// If the device is a MIG device (i.e. a compute instance) these are 0
func (d LibDevice) GetCudaComputeCapability() (int, int, error) {
	if d.isMigDevice {
		return 0, 0, nil
	}

	major, minor, ret := d.device.GetCudaComputeCapability()
	if ret != nvml.SUCCESS {
		return 0, 0, errorString(ret)
	}

	return major, minor, nil
}

// GetUUID returns the UUID of the CUDA device
func (d LibDevice) GetUUID() (string, error) {
	uuid, ret := d.device.GetUUID()
	if ret != nvml.SUCCESS {
		return "", errorString(ret)
	}

	return uuid, nil
}

// GetName returns the device name / model.
func (d LibDevice) GetName() (string, error) {
	name, ret := d.device.GetName()
	if ret != nvml.SUCCESS {
		return "", errorString(ret)
	}

	return name, nil
}

// GetMemoryInfo returns the total and available memory for a device
func (d LibDevice) GetMemoryInfo() (Memory, error) {
	info, ret := d.device.GetMemoryInfo()
	if ret != nvml.SUCCESS {
		return Memory{}, errorString(ret)
	}

	return Memory(info), nil
}

func errorString(r nvml.Return) error {
	if r == nvml.SUCCESS {
		return nil
	}
	return fmt.Errorf("NVML error: %v", nvml.ErrorString(r))
}

// GetArchFamily returns the architecture family string for the device
func (d LibDevice) GetArchFamily() (string, error) {
	computeMajor, computeMinor, err := d.GetCudaComputeCapability()
	if err != nil {
		return "", err
	}
	return getArchFamily(computeMajor, computeMinor), nil
}
