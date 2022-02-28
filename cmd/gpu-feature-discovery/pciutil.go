package main

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"
)

// NvidiaPCI interface allows us to get a list of all NVIDIA PCI devices
type NvidiaPCI interface {
	Devices() ([]*PCIDevice, error)
}

// PCIDevice represents a single PCI device
type PCIDevice struct {
	Path    string
	Address string
	Class   string
	Vendor  string
	Config  []byte
}

const (
	// PciDevicesRoot represents base path for all pci devices under sysfs
	PciDevicesRoot = "/sys/bus/pci/devices"
	// PciStatusByte indicates status byte
	PciStatusByte = 0x06
	// PciStatusCapabilityList indicates if capability list is supported
	PciStatusCapabilityList = 0x10
	// PciCapabilityList indicates offset of first capability list entry
	PciCapabilityList = 0x34
	// PciCapabilityListID indicates offset for capability id
	PciCapabilityListID = 0
	// PciCapabilityListNext indicates offset for next capability in the list
	PciCapabilityListNext = 1
	// PciCapabilityLength indicates offset for capability length
	PciCapabilityLength = 2
	// PciCapabilityVendorSpecificID indicates PCI vendor specific capability id
	PciCapabilityVendorSpecificID = 0x09
	// PciNvidiaVendorID represents PCI vendor id for Nvidia
	PciNvidiaVendorID = "0x10de"
)

// NvidiaPCILib implements the NvidiaPCI interface
type NvidiaPCILib struct{}

// NewNvidiaPCILib returns an instance of NvidiaPCILib implementing the NvidiaPCI interface
func NewNvidiaPCILib() NvidiaPCI {
	return &NvidiaPCILib{}
}

// Devices returns all PCI devices on the system
func (p *NvidiaPCILib) Devices() ([]*PCIDevice, error) {
	deviceDirs, err := ioutil.ReadDir(PciDevicesRoot)
	if err != nil {
		return nil, fmt.Errorf("unable to read PCI bus devices: %v", err)
	}

	var devices []*PCIDevice
	for _, deviceDir := range deviceDirs {
		devicePath := path.Join(PciDevicesRoot, deviceDir.Name())
		address := deviceDir.Name()

		vendor, err := ioutil.ReadFile(path.Join(devicePath, "vendor"))
		if err != nil {
			return nil, fmt.Errorf("unable to read PCI device vendor id for %s: %v", address, err)
		}

		if strings.TrimSpace(string(vendor)) != PciNvidiaVendorID {
			continue
		}

		class, err := ioutil.ReadFile(path.Join(devicePath, "class"))
		if err != nil {
			return nil, fmt.Errorf("unable to read PCI device class for %s: %v", address, err)
		}

		config, err := ioutil.ReadFile(path.Join(devicePath, "config"))
		if err != nil {
			return nil, fmt.Errorf("unable to read PCI configuration space for %s: %v", address, err)
		}

		device := &PCIDevice{
			Path:    devicePath,
			Address: address,
			Vendor:  strings.TrimSpace(string(vendor)),
			Class:   string(class)[0:4],
			Config:  config,
		}

		devices = append(devices, device)
	}

	return devices, nil
}

// GetVendorSpecificCapability returns the vendor specific capability from configuration space
func (d *PCIDevice) GetVendorSpecificCapability() ([]byte, error) {
	if len(d.Config) < 256 {
		return nil, fmt.Errorf("entire PCI configuration is not read for device %s. Please run GFD with privileged mode to read complete PCI configuration data", d.Address)
	}

	if d.Config[PciStatusByte]&PciStatusCapabilityList == 0 {
		return nil, nil
	}

	var visited [256]byte
	pos := int(GetByte(d.Config, PciCapabilityList))
	for pos != 0 {
		id := int(GetByte(d.Config, pos+PciCapabilityListID))
		next := int(GetByte(d.Config, pos+PciCapabilityListNext))
		length := int(GetByte(d.Config, pos+PciCapabilityLength))

		if visited[pos] != 0 {
			// chain looped
			break
		}
		if id == 0xff {
			// chain broken
			break
		}
		if id == PciCapabilityVendorSpecificID {
			capability := d.Config[pos+PciCapabilityListID : pos+PciCapabilityListID+length]
			return capability, nil
		}

		visited[pos]++
		pos = next
	}

	return nil, nil
}

// GetByte returns a single byte of data at specified position
func GetByte(buffer []byte, pos int) uint8 {
	return uint8(buffer[pos])
}

// GetWord returns 2 bytes of data from specified position
func GetWord(buffer []byte, pos int) uint16 {
	return uint16(buffer[pos]) | (uint16(buffer[pos+1]) << 8)
}

// GetLong returns 4 bytes of data from specified position
func GetLong(buffer []byte, pos int) uint32 {
	return uint32(buffer[pos]) |
		uint32(buffer[pos+1])<<8 |
		uint32(buffer[pos+2])<<16 |
		uint32(buffer[pos+3])<<24
}
