package vgpu

import (
	"fmt"
	"log"

	"github.com/NVIDIA/gpu-feature-discovery/pkg/pciutil"
)

// NvidiaMockVGPU represents implementation of Nvidia vGPU interfaces
type NvidiaMockVGPU struct {
	pci     pciutil.MockPCI
	addVGPU bool // indicates to add vgpu mock device to devices list
}

// NewNvidiaMockVGPU returns an instance of  VGPU interface for Nvidia devices
func NewNvidiaMockVGPU(addVirtualDevice bool) NvidiaMockVGPU {
	return NvidiaMockVGPU{pci: pciutil.MockPCI{AddVGPU: addVirtualDevice}}
}

// GetAllVGPUDevices returns all vGPU devices attached to the guest
func (v NvidiaMockVGPU) GetAllVGPUDevices() ([]*pciutil.NvidiaPCIDevice, error) {
	var vGPUDevices []*pciutil.NvidiaPCIDevice
	err := v.pci.GetPCIDevices()
	if err != nil {
		return nil, fmt.Errorf("Unable to find PCI devices by nvidia vendor id 0x10de : %v", err)
	}

	for _, device := range v.pci.Devices {
		// fetch config
		err := device.ReadConfig()
		if err != nil {
			return nil, fmt.Errorf("Unable to read PCI configuration for %s: %v", device.Address, err)
		}
		// fetch vendor capabilities
		err = device.GetVendorCapabilities()
		if err != nil {
			return nil, fmt.Errorf("Unable to read vendor capabilities for %s: %v", device.Address, err)
		}
		if vgpu := IsVGPUDevice(device); vgpu {
			vGPUDevices = append(vGPUDevices, device)
		}
	}
	return vGPUDevices, nil
}

// IsVGPUDevicePresent returns true if a guest is attached with a vGPU device
func (v NvidiaMockVGPU) IsVGPUDevicePresent() (bool, error) {
	devices, err := v.GetAllVGPUDevices()
	if err != nil {
		return false, err
	}
	if len(devices) > 0 {
		log.Printf("Found %d vGPU devices", len(devices))
		return true, nil
	}
	return false, nil
}
