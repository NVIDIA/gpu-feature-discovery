package main

import (
	"fmt"
	"strings"
)

// VGPU interface allows us to get a list of vGPU specific PCI devices
type VGPU interface {
	Devices() ([]*VGPUDevice, error)
}

// VGPUDevice is just an alias to a PCIDevice
type VGPUDevice struct {
	pci            *PCIDevice
	vGPUCapability []byte
}

// VGPUInfo represents vGPU driver info running on underlying hypervisor host.
type VGPUInfo struct {
	HostDriverVersion string
	HostDriverBranch  string
}

const (
	// VGPUCapabilityRecordStart indicates offset of beginning vGPU capability record
	VGPUCapabilityRecordStart = 5
	// HostDriverVersionLength indicates max length of driver version
	HostDriverVersionLength = 10
	// HostDriverBranchLength indicates max length of driver branch
	HostDriverBranchLength = 10
)

// VGPULib implements the NvidiaVGPU interface
type VGPULib struct {
	pci NvidiaPCI
}

// NewVGPULib returns an instance of VGPULib implementing the VGPU interface
func NewVGPULib(pci NvidiaPCI) VGPU {
	return &VGPULib{pci: pci}
}

// Devices returns all vGPU devices attached to the guest
func (v *VGPULib) Devices() ([]*VGPUDevice, error) {
	pciDevices, err := v.pci.Devices()
	if err != nil {
		return nil, fmt.Errorf("Error getting NVIDIA specific PCI devices: %v", err)
	}

	var vgpus []*VGPUDevice
	for _, device := range pciDevices {
		capability, err := device.GetVendorSpecificCapability()
		if err != nil {
			return nil, fmt.Errorf("Unable to read vendor specific capability for %s: %v", device.Address, err)
		}
		if capability == nil {
			continue
		}
		if exists := v.IsVGPUDevice(capability); exists {
			vgpu := &VGPUDevice{
				pci:            device,
				vGPUCapability: capability,
			}
			vgpus = append(vgpus, vgpu)
		}
	}
	return vgpus, nil
}

// IsVGPUDevice returns true if the device is of type vGPU
func (v *VGPULib) IsVGPUDevice(capability []byte) bool {
	if len(capability) < 5 {
		return false
	}
	// check for vGPU signature, 0x56, 0x46 i.e "VF"
	if capability[3] != 0x56 {
		return false
	}
	if capability[4] != 0x46 {
		return false
	}
	return true
}

// GetInfo returns information about vGPU manager running on the underlying hypervisor host
func (d *VGPUDevice) GetInfo() (*VGPUInfo, error) {
	if len(d.vGPUCapability) == 0 {
		return nil, fmt.Errorf("Vendor capability record is not populated for device %s", d.pci.Address)
	}

	// traverse vGPU vendor capability records until host driver version record(id: 0) is found
	var hostDriverVersion string
	var hostDriverBranch string
	foundDriverVersionRecord := false
	pos := VGPUCapabilityRecordStart
	record := GetByte(d.vGPUCapability, VGPUCapabilityRecordStart)
	for record != 0 && pos < len(d.vGPUCapability) {
		// find next record
		recordLength := GetByte(d.vGPUCapability, pos+1)
		pos = pos + int(recordLength)
		record = GetByte(d.vGPUCapability, pos)
	}

	if record == 0 && pos+2+HostDriverVersionLength+HostDriverBranchLength <= len(d.vGPUCapability) {
		foundDriverVersionRecord = true
		// found vGPU host driver version record type
		// initialized at record data byte, i.e pos + 1(record id byte) + 1(record lengh byte)
		i := pos + 2
		// 10 bytes of driver version
		for ; i < pos+2+HostDriverVersionLength; i++ {
			hostDriverVersion += string(GetByte(d.vGPUCapability, i))
		}
		hostDriverVersion = strings.Trim(hostDriverVersion, "\x00")
		// 10 bytes of driver branch
		for ; i < pos+2+HostDriverVersionLength+HostDriverBranchLength; i++ {
			hostDriverBranch += string(GetByte(d.vGPUCapability, i))
		}
		hostDriverBranch = strings.Trim(hostDriverBranch, "\x00")
	}

	if !foundDriverVersionRecord {
		return nil, fmt.Errorf("Cannot find driver version record in vendor specific capability for device %s", d.pci.Address)
	}

	info := &VGPUInfo{
		HostDriverVersion: hostDriverVersion,
		HostDriverBranch:  hostDriverBranch,
	}

	return info, nil
}
