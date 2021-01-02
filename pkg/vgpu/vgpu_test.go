package vgpu

import (
	"testing"

	"github.com/NVIDIA/gpu-feature-discovery/pkg/pciutil"
	"github.com/stretchr/testify/require"
)

func NewMockPCIDevice() pciutil.PCIDevice {
	return pciutil.PCIDevice{Address: "0000:0b:00.0", Vendor: "0x10de", FullPath: ".", Class: "300"}
}

func TestIsVGPUDevice(t *testing.T) {
	device := NewMockPCIDevice()
	err := device.ReadConfig()
	require.NoError(t, err, "Reading configuration space")
	require.NotEmpty(t, device.Config, "Device Configuration data")
	require.Equal(t, len(device.Config), 256, "Device configuration data length")

	err = device.GetVendorCapabilities()
	require.NoError(t, err, "Get Vendor capability records")
	require.NotEmpty(t, device.VendorCapability, "Vendor capability record")
	require.Equal(t, device.VendorCapability[0], uint8(9), "Vendor capability id")

	require.True(t, IsVGPUDevice(&device), "Is vGPU device")
}

func TestGetHostDriverVersionAndBranch(t *testing.T) {
	device := NewMockPCIDevice()
	err := device.ReadConfig()
	require.NoError(t, err, "Reading configuration space")
	require.NotEmpty(t, device.Config, "Device Configuration data")
	require.Equal(t, len(device.Config), 256, "Device configuration data length")

	err = device.GetVendorCapabilities()
	require.NoError(t, err, "Get Vendor capability records")
	require.NotEmpty(t, device.VendorCapability, "Vendor capability record")
	require.Equal(t, device.VendorCapability[0], uint8(9), "Vendor capability id")

	driverInfo, err := GetHostDriverInfo(&device)
	require.NoError(t, err, "Get host driver version and branch")
	require.NotNil(t, driverInfo, "Host driver info")
	require.Equal(t, "460.16", driverInfo.Driver, "Host driver version")
	require.Equal(t, "r460_00", driverInfo.Branch, "Host driver branch")
}
