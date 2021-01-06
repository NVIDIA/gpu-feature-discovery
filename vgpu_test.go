package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// MockVGPU represents mock of VGPU interface
type MockVGPU struct {
	devices []*VGPUDevice
}

// Devices returns VGPU devices with mocked data
func (p *MockVGPU) Devices() ([]*VGPUDevice, error) {
	return p.devices, nil
}

// NewMockVGPU initializes and returns mock VGPU interface type
func NewMockVGPU() VGPU {
	return NewVGPULib(NewMockNvidiaPCI())
}

func TestIsVGPUDevice(t *testing.T) {
	mockVGPU := NewMockVGPU().(*VGPULib)
	devices, _ := mockVGPU.pci.Devices()
	for _, device := range devices {
		// check for vendor id
		require.Equal(t, "0x10de", fmt.Sprintf("0x%x", GetWord(device.Config, 0)), "Nvidia PCI Vendor ID")
		// check for vendor capability records
		capability, err := device.GetVendorSpecificCapability()
		require.NoError(t, err, "Get vendor capabilities from configuration space")
		require.NotZero(t, len(capability), "Vendor capability record")
		if device.Address == "passthrough" {
			require.False(t, mockVGPU.IsVGPUDevice(capability), "Is not a virtual GPU device")
			require.Equal(t, 20, len(capability), "Vendor capability length for passthrough device")
		}
		if device.Address == "vgpu" {
			require.Equal(t, 27, len(capability), "Vendor capability length for vgpu device")
			require.Equal(t, uint8(9), GetByte(capability, 0), "Vendor capability ID")
		}
	}
}

func TestVGPUGetInfo(t *testing.T) {
	devices, _ := NewMockVGPU().Devices()
	for _, device := range devices {
		if device.pci.Address == "vgpu" {
			require.NotEmpty(t, device.pci.Config, "Device Configuration data")
			require.Equal(t, len(device.pci.Config), 256, "Device configuration data length")

			require.NotEmpty(t, device.vGPUCapability, "Vendor capability record")
			require.Equal(t, device.vGPUCapability[0], uint8(9), "Vendor capability id")

			info, err := device.GetInfo()
			require.NoError(t, err, "Get host driver version and branch")
			require.NotNil(t, info, "Host driver info")
			require.Equal(t, "460.16", info.HostDriverVersion, "Host driver version")
			require.Equal(t, "r460_00", info.HostDriverBranch, "Host driver branch")
		}
	}
}
