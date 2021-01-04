package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetVendorCapabilities(t *testing.T) {
	mockPCI := NewMockPCI(false)
	mockPCI.GetPCIDevices()
	for _, device := range mockPCI.Devices {
		// check for vendor id
		require.Equal(t, "0x10de", fmt.Sprintf("0x%x", device.GetConfigWord(0, device.Config)), "Nvidia PCI Vendor ID")
		// check for vendor capability records
		err := device.GetVendorCapabilities()
		require.NoError(t, err, "Get vendor capabilities from configuration space")
		require.NotZero(t, len(device.VendorCapability), "Vendor capability record")
		if device.Address == "passthrough" {
			require.Equal(t, 20, len(device.VendorCapability), "Vendor capability length for passthrough device")
		}
	}
}
