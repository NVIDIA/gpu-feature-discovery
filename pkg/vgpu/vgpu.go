package vgpu

import (
	"fmt"
	"log"
	"strings"

	"github.com/NVIDIA/gpu-feature-discovery/pkg/pciutil"
)

const (
	// NvidiaVendorID indicates vendor id for Nvidia PCI devices
	NvidiaVendorID = "0x10de"
	// VGPUCapabilityRecordStart indicates offset of beginning vGPU capability record
	VGPUCapabilityRecordStart = 5
	// HostDriverVersionLength indicates max length of driver version
	HostDriverVersionLength = 10
	// HostDriverBranchLength indicates max length of driver branch
	HostDriverBranchLength = 10
)

// IsVGPUDevicePresent returns true if a guest is attached with a vGPU device
func IsVGPUDevicePresent() (bool, error) {
	devices, err := GetAllVGPUDevices()
	if err != nil {
		return false, err
	}
	if len(devices) > 0 {
		log.Printf("Found %d vGPU devices", len(devices))
		return true, nil
	}
	return false, nil
}

// IsVGPUDevice returns true if the device is of type vGPU
func IsVGPUDevice(d *pciutil.PCIDevice) bool {
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

// GetAllVGPUDevices returns all vGPU devices attached to the guest
func GetAllVGPUDevices() ([]pciutil.PCIDevice, error) {
	log.Printf(">>>>> GetAllVGPUDevices")
	defer log.Printf("<<<<< GetAllVGPUDevices")
	var vGPUDevices []pciutil.PCIDevice
	devices, err := pciutil.GetDevicesByVendorID(NvidiaVendorID)
	if err != nil {
		return nil, fmt.Errorf("Unable to find PCI devices by nvidia vendor id 0x10de : %v", err)
	}

	for _, device := range devices {
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
		if vgpu := IsVGPUDevice(&device); vgpu {
			vGPUDevices = append(vGPUDevices, device)
		}
	}
	return vGPUDevices, nil
}

// GetHostDriverVersionAndBranch returns driver version and branch of vGPU manager running on the underlying hypervisor host
func GetHostDriverVersionAndBranch(d *pciutil.PCIDevice) (string, string, error) {
	if len(d.VendorCapability) == 0 {
		return "", "", fmt.Errorf("Vendor capability record is not populated for device %s", d.Address)
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
		return "", "", fmt.Errorf("Cannot find driver version record in vendor specific capability for device %s", d.Address)
	}
	log.Printf("found host driver version %s and branch %s for device %s", hostDriverVersion, hostDriverBranch, d.Address)
	return hostDriverVersion, hostDriverBranch, nil
}
