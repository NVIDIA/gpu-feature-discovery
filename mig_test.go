// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package main

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
	"github.com/stretchr/testify/require"
)

func TestMigStrategySingle(t *testing.T) {
	nvmlMock := NewTestNvmlMock()
	// create VGPU mock library with empty vgpu devices
	vgpuMock := NewTestVGPUMock(false)
	nvmlMock.devices[0].migEnabled = true
	nvmlMock.devices[0].migDevices = []NvmlMockDevice{
		NvmlMockDevice{
			instance: &nvml.Device{
				Model: nvmlMock.devices[0].Instance().Model,
			},
			attributes: &nvml.DeviceAttributes{
				GpuInstanceSliceCount: 3,
				MemorySizeMB:          20000,
			},
		},
		NvmlMockDevice{
			instance: &nvml.Device{
				Model: nvmlMock.devices[0].Instance().Model,
			},
			attributes: &nvml.DeviceAttributes{
				GpuInstanceSliceCount: 3,
				MemorySizeMB:          20000,
			},
		},
	}

	conf := Conf{true, "single", "./gfd-test-mig-single", time.Second}

	MachineTypePath = "/tmp/machine-type"
	machineType := []byte("product-name\n")
	err := ioutil.WriteFile("/tmp/machine-type", machineType, 0644)
	require.NoError(t, err, "Write machine type mock file")

	defer func() {
		err = os.Remove(MachineTypePath)
		require.NoError(t, err, "Removing machine type mock file")
	}()

	err = run(nvmlMock, vgpuMock, conf)
	require.NoError(t, err, "Error from run function")

	outFile, err := os.Open(conf.OutputFilePath)
	require.NoError(t, err, "Opening output file")

	defer func() {
		err = outFile.Close()
		require.NoError(t, err, "Closing output file")
		err = os.Remove(conf.OutputFilePath)
		require.NoError(t, err, "Removing output file")
	}()

	output, err := ioutil.ReadAll(outFile)
	require.NoError(t, err, "Reading output file")

	err = checkResult(output, "tests/expected-output-mig-single.txt")
	require.NoError(t, err, "Checking result")

	labels, err := buildLabelMapFromOutput(output)
	require.NoError(t, err, "Building map of labels from output file")

	require.Equal(t, labels["nvidia.com/mig.strategy"], "single", "Incorrect label")
	require.Equal(t, labels["nvidia.com/gpu.count"], "2", "Incorrect label")
	require.Equal(t, labels["nvidia.com/gpu.product"], "MOCKMODEL-MIG-3g.20gb", "Incorrect label")
	require.Equal(t, labels["nvidia.com/gpu.memory"], "20000", "Incorrect label")
}

func TestMigStrategyMixed(t *testing.T) {
	nvmlMock := NewTestNvmlMock()
	// create VGPU mock library with empty vgpu devices
	vgpuMock := NewTestVGPUMock(false)

	nvmlMock.devices[0].migEnabled = true
	nvmlMock.devices[0].migDevices = []NvmlMockDevice{
		NvmlMockDevice{
			attributes: &nvml.DeviceAttributes{
				GpuInstanceSliceCount: 3,
				MemorySizeMB:          20000,
			},
		},
		NvmlMockDevice{
			attributes: &nvml.DeviceAttributes{
				GpuInstanceSliceCount: 1,
				MemorySizeMB:          5000,
			},
		},
	}

	conf := Conf{true, "mixed", "./gfd-test-mig-mixed", time.Second}

	MachineTypePath = "/tmp/machine-type"
	machineType := []byte("product-name\n")
	err := ioutil.WriteFile("/tmp/machine-type", machineType, 0644)
	require.NoError(t, err, "Write machine type mock file")

	defer func() {
		err = os.Remove(MachineTypePath)
		require.NoError(t, err, "Removing machine type mock file")
	}()

	err = run(nvmlMock, vgpuMock, conf)
	require.NoError(t, err, "Error from run function")

	outFile, err := os.Open(conf.OutputFilePath)
	require.NoError(t, err, "Opening output file")

	defer func() {
		err = outFile.Close()
		require.NoError(t, err, "Closing output file")
		err = os.Remove(conf.OutputFilePath)
		require.NoError(t, err, "Removing output file")
	}()

	output, err := ioutil.ReadAll(outFile)
	require.NoError(t, err, "Reading output file")

	err = checkResult(output, "tests/expected-output-mig-mixed.txt")
	require.NoError(t, err, "Checking result")

	labels, err := buildLabelMapFromOutput(output)
	require.NoError(t, err, "Building map of labels from output file")

	require.Equal(t, labels["nvidia.com/mig.strategy"], "mixed", "Incorrect label")
	require.Equal(t, labels["nvidia.com/gpu.count"], "1", "Incorrect label")
	require.Equal(t, labels["nvidia.com/gpu.product"], "MOCKMODEL", "Incorrect label")
	require.Equal(t, labels["nvidia.com/gpu.memory"], "128", "Incorrect label")
	require.Contains(t, labels, "nvidia.com/mig-3g.20gb.count", "Missing label")
	require.Contains(t, labels, "nvidia.com/mig-1g.5gb.count", "Missing label")
}
