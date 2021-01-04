package main

import (
	"fmt"
	"log"
	"strings"
)

const (
	// VGPUCapabilityRecordStart indicates offset of beginning vGPU capability record
	VGPUCapabilityRecordStart = 5
	// HostDriverVersionLength indicates max length of driver version
	HostDriverVersionLength = 10
	// HostDriverBranchLength indicates max length of driver branch
	HostDriverBranchLength = 10
)

// VirtualGPU represents vGPU interface
type VirtualGPU interface {
	IsVGPUDevicePresent() (bool, error)
	GetAllVGPUDevices() ([]*NvidiaPCIDevice, error)
}

// NvidiaVGPU represents implementation of Nvidia vGPU interfaces
type NvidiaVGPU struct {
	pci NvidiaPCI
}

// NewNvidiaVGPU returns an instance of  VGPU interface for Nvidia devices
func NewNvidiaVGPU() NvidiaVGPU {
	return NvidiaVGPU{pci: NvidiaPCI{}}
}

// HostDriverInfo represents vGPU driver info running on underlying hypervisor host.
type HostDriverInfo struct {
	Version string
	Branch  string
}

// IsVGPUDevicePresent returns true if a guest is attached with a vGPU device
func (v NvidiaVGPU) IsVGPUDevicePresent() (bool, error) {
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

// GetAllVGPUDevices returns all vGPU devices attached to the guest
func (v NvidiaVGPU) GetAllVGPUDevices() ([]*NvidiaPCIDevice, error) {
	var vGPUDevices []*NvidiaPCIDevice
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

// IsVGPUDevice returns true if the device is of type vGPU
func IsVGPUDevice(d *NvidiaPCIDevice) bool {
	if len(d.VendorCapability) < 5 {
		return false
	}
	// check for vGPU signature, 0x56, 0x46 i.e "VF"
	log.Println(d.VendorCapability[3])
	log.Println(d.VendorCapability[4])
	if d.VendorCapability[3] == 0x56 && d.VendorCapability[4] == 0x46 {
		log.Printf("Found vGPU device %s", d.Address)
		return true
	}
	return false
}

// GetHostDriverInfo returns information about vGPU manager running on the underlying hypervisor host
func GetHostDriverInfo(d *NvidiaPCIDevice) (*HostDriverInfo, error) {
	if len(d.VendorCapability) == 0 {
		return nil, fmt.Errorf("Vendor capability record is not populated for device %s", d.Address)
	}
	var hostDriverVersion string
	var hostDriverBranch string
	foundDriverVersionRecord := false
	// traverse vGPU vendor capability records
	pos := uint8(VGPUCapabilityRecordStart)
	record := d.GetByte(VGPUCapabilityRecordStart, d.VendorCapability)
	// traverse until host driver version record(id: 0) is found
	for record != 0 && pos < uint8(len(d.VendorCapability)) {
		// find next record
		recordLength := d.GetByte(pos+1, d.VendorCapability)
		pos = pos + recordLength
		record = d.GetByte(pos, d.VendorCapability)
	}
	if record == 0 && pos+2+HostDriverVersionLength+HostDriverBranchLength <= uint8(len(d.VendorCapability)) {
		foundDriverVersionRecord = true
		// found vGPU host driver version record type
		// initialized at record data byte, i.e pos + 1(record id byte) + 1(record lengh byte)
		i := pos + 2
		// 10 bytes of driver version
		for ; i < pos+2+HostDriverVersionLength; i++ {
			hostDriverVersion += string(d.GetByte(i, d.VendorCapability))
		}
		hostDriverVersion = strings.Trim(hostDriverVersion, "\x00")
		// 10 bytes of driver branch
		for ; i < pos+2+HostDriverVersionLength+HostDriverBranchLength; i++ {
			hostDriverBranch += string(d.GetByte(i, d.VendorCapability))
		}
		hostDriverBranch = strings.Trim(hostDriverBranch, "\x00")
	}

	if !foundDriverVersionRecord {
		return nil, fmt.Errorf("Cannot find driver version record in vendor specific capability for device %s", d.Address)
	}
	log.Printf("found host driver version %s and branch %s for device %s", hostDriverVersion, hostDriverBranch, d.Address)
	return &HostDriverInfo{Version: hostDriverVersion, Branch: hostDriverBranch}, nil
}
