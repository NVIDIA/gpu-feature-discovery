package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"strings"
)

// PCI represents interactions with PCI
type PCI interface {
	GetPCIDevices() error
}

// PCIDevice represents interactions with PCI.device
type PCIDevice interface {
	ReadConfig() error
	GetVendorCapabilities() error
	GetByte(pos uint8, config []byte) uint8
	GetConfigWord(pos uint8, config []byte) uint16
	GetConfigLong(pos uint8, config []byte) uint32
}

// NvidiaPCI represents PCI interface implementation for Nvidia PCI devices
type NvidiaPCI struct {
	Devices []*NvidiaPCIDevice
}

// NvidiaPCIDevice represents Nvidia PCI device
type NvidiaPCIDevice struct {
	Address          string
	Class            string
	Vendor           string
	Device           string
	FullPath         string
	Config           []byte
	VendorCapability []byte
}

const (
	// SysBusPCIDevices represents base path for all pci devices under sysfs
	SysBusPCIDevices = "/sys/bus/pci/devices"
	// NvidiaVendorID represents PCI vendor id for Nvidia
	NvidiaVendorID = "0x10de"
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
	// PciCapabilityIDVendor indicates PCI vendor capability id
	PciCapabilityIDVendor = 0x09
)

// GetPCIDevices returns all PCI devices on the system
func (p *NvidiaPCI) GetPCIDevices() error {
	devices, err := ioutil.ReadDir(SysBusPCIDevices)
	if err != nil {
		return fmt.Errorf("Unable to read PCI bus devices: %v", err)
	}
	for _, device := range devices {
		// read basic information for each device
		vendor, err := ioutil.ReadFile(path.Join(SysBusPCIDevices, device.Name(), "vendor"))
		if err != nil {
			return fmt.Errorf("Unable to read PCI device vendor id for %s: %v", device.Name(), err)
		}
		// ignore PCI devices other than Nvidia
		if strings.TrimSpace(string(vendor)) != NvidiaVendorID {
			continue
		}

		class, err := ioutil.ReadFile(path.Join(SysBusPCIDevices, device.Name(), "class"))
		if err != nil {
			return fmt.Errorf("Unable to read PCI device class for %s: %v", device.Name(), err)
		}
		p.Devices = append(p.Devices, &NvidiaPCIDevice{Address: device.Name(), Vendor: strings.TrimSpace(string(vendor)), Class: string(class)[0:4], FullPath: path.Join(SysBusPCIDevices, device.Name())})
	}
	return nil
}

// ReadConfig reads PCI configuration space of device
func (d *NvidiaPCIDevice) ReadConfig() error {
	if len(d.Config) != 0 {
		// config is already loaded
		return nil
	}
	config, err := ioutil.ReadFile(path.Join(d.FullPath, "config"))
	if err != nil {
		return fmt.Errorf("Unable to read PCI configuration space: %v", err)
	}
	d.Config = config
	return nil
}

// GetVendorCapabilities returns vendor specific capabilities from configuration space
func (d *NvidiaPCIDevice) GetVendorCapabilities() error {
	if d.Config[PciStatusByte]&PciStatusCapabilityList == 0 {
		// capability list is not supported
		log.Printf("Capability records are not supported for device %s", d.Address)
		return nil
	}

	// check if entire configuration data is read from sysfs
	if len(d.Config) < 256 {
		return fmt.Errorf("Entire PCI configuration is not read for device %s. Please run GFD with privileged mode to read complete PCI configuration data", d.Address)
	}

	var visited [256]byte
	pos := d.GetByte(PciCapabilityList, d.Config)
	for pos != 0 {
		id := uint8(0)
		next := uint8(0)
		length := uint8(0)

		id = d.GetByte(pos+PciCapabilityListID, d.Config)
		next = d.GetByte(pos+PciCapabilityListNext, d.Config)
		length = d.GetByte(pos+PciCapabilityLength, d.Config)

		if visited[pos] != 0 {
			// chain looped
			log.Println("chain looped, exiting")
			break
		}
		if id == 0xff {
			// chain broken
			log.Println("chain broken, exiting")
			break
		}
		if id == PciCapabilityIDVendor {
			// add capability to the vendor cap list
			log.Printf("found vendor specific capability for %s", d.Address)
			d.VendorCapability = d.Config[pos+PciCapabilityListID : pos+PciCapabilityListID+length]
			log.Println(hex.Dump(d.VendorCapability))
			log.Println(len(d.VendorCapability))
		}
		visited[pos]++
		pos = next
	}
	return nil
}

// GetByte returns a single byte of config data at specified position
func (d *NvidiaPCIDevice) GetByte(pos uint8, config []byte) uint8 {
	return uint8(config[pos])
}

// GetConfigWord returns 2 bytes of config data from specified position
func (d *NvidiaPCIDevice) GetConfigWord(pos uint8, config []byte) uint16 {
	return uint16(config[pos]) | (uint16(config[pos+1]) << 8)
}

// GetConfigLong returns 4 bytes of config data from specified position
func (d *NvidiaPCIDevice) GetConfigLong(pos uint8, config []byte) uint32 {
	return uint32(config[pos]) |
		uint32(config[pos+1])<<8 |
		uint32(config[pos+2])<<16 |
		uint32(config[pos+3])<<24
}
