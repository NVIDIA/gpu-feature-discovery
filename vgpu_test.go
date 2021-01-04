package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsVGPUDevice(t *testing.T) {
	mockPCI := NewMockPCI(true)
	mockPCI.GetPCIDevices()

	for _, device := range mockPCI.Devices {
		// check for vendor id
		require.Equal(t, "0x10de", fmt.Sprintf("0x%x", device.GetConfigWord(0, device.Config)), "Nvidia PCI Vendor ID")
		// check for vendor capability records
		err := device.GetVendorCapabilities()
		require.NoError(t, err, "Get vendor capabilities from configuration space")
		require.NotZero(t, len(device.VendorCapability), "Vendor capability record")
		if device.Address == "passthrough" {
			require.False(t, IsVGPUDevice(device), "Is not a virtual GPU device")
			require.Equal(t, 20, len(device.VendorCapability), "Vendor capability length for passthrough device")
		}
		if device.Address == "vgpu" {
			require.Equal(t, 27, len(device.VendorCapability), "Vendor capability length for vgpu device")
			require.Equal(t, uint8(9), device.GetByte(0, device.VendorCapability), "Vendor capability ID")
		}
	}
}

func TestGetHostDriverVersionAndBranch(t *testing.T) {
	mockPCI := NewMockPCI(true)
	mockPCI.GetPCIDevices()

	for _, device := range mockPCI.Devices {
		if device.Address == "vgpu" {
			err := device.ReadConfig()
			require.NoError(t, err, "Reading configuration space")
			require.NotEmpty(t, device.Config, "Device Configuration data")
			require.Equal(t, len(device.Config), 256, "Device configuration data length")

			err = device.GetVendorCapabilities()
			require.NoError(t, err, "Get Vendor capability records")
			require.NotEmpty(t, device.VendorCapability, "Vendor capability record")
			require.Equal(t, device.VendorCapability[0], uint8(9), "Vendor capability id")

			driverInfo, err := GetHostDriverInfo(device)
			require.NoError(t, err, "Get host driver version and branch")
			require.NotNil(t, driverInfo, "Host driver info")
			require.Equal(t, "460.16", driverInfo.Version, "Host driver version")
			require.Equal(t, "r460_00", driverInfo.Branch, "Host driver branch")
		}
	}
}
