// Copyright (c) 2019, NVIDIA CORPORATION. All rights reserved.

package main

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/NVIDIA/gpu-feature-discovery/internal/nvml"
	config "github.com/NVIDIA/k8s-device-plugin/api/config/v1"
	"github.com/stretchr/testify/require"
)

func TestMigStrategyNone(t *testing.T) {
	nvmlMock := NewTestNvmlMock()
	// create VGPU mock library with empty vgpu devices
	vgpuMock := NewTestVGPUMock()

	nvmlMock.Devices[0].MigEnabled = true
	nvmlMock.Devices[0].MigDevices = []nvml.NvmlMockDevice{
		{
			Model: "MOCKMODEL",
			Attributes: &nvml.DeviceAttributes{
				GpuInstanceSliceCount: 3,
				MemorySizeMB:          20000,
			},
		},
		{
			Model: "MOCKMODEL",
			Attributes: &nvml.DeviceAttributes{
				GpuInstanceSliceCount: 3,
				MemorySizeMB:          20000,
			},
		},
	}

	conf := &config.Config{
		Flags: config.Flags{
			CommandLineFlags: config.CommandLineFlags{
				MigStrategy:     "none",
				FailOnInitError: true,
				GFD: config.GFDCommandLineFlags{
					Oneshot:       true,
					OutputFile:    "./gfd-test-mig-none",
					SleepInterval: time.Second,
					NoTimestamp:   false,
				},
			},
		},
	}

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

	outFile, err := os.Open(conf.Flags.GFD.OutputFile)
	require.NoError(t, err, "Opening output file")

	defer func() {
		err = outFile.Close()
		require.NoError(t, err, "Closing output file")
		err = os.Remove(conf.Flags.GFD.OutputFile)
		require.NoError(t, err, "Removing output file")
	}()

	output, err := ioutil.ReadAll(outFile)
	require.NoError(t, err, "Reading output file")

	err = checkResult(output, cfg.Path("tests/expected-output-mig-none.txt"), false)
	require.NoError(t, err, "Checking result")

	labels, err := buildLabelMapFromOutput(output)
	require.NoError(t, err, "Building map of labels from output file")

	require.Equal(t, labels["nvidia.com/gpu.count"], "1", "Incorrect label")
	require.Equal(t, labels["nvidia.com/gpu.product"], "MOCKMODEL", "Incorrect label")
	require.Equal(t, labels["nvidia.com/gpu.memory"], "128", "Incorrect label")
}

func TestMigStrategySingleForNoMigDevices(t *testing.T) {
	nvmlMock := NewTestNvmlMock()
	// create VGPU mock library with empty vgpu devices
	vgpuMock := NewTestVGPUMock()

	conf := &config.Config{
		Flags: config.Flags{
			CommandLineFlags: config.CommandLineFlags{
				MigStrategy:     "single",
				FailOnInitError: true,
				GFD: config.GFDCommandLineFlags{
					Oneshot:       true,
					OutputFile:    "./gfd-test-mig-single-no-mig",
					SleepInterval: time.Second,
					NoTimestamp:   false,
				},
			},
		},
	}

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

	outFile, err := os.Open(conf.Flags.GFD.OutputFile)
	require.NoError(t, err, "Opening output file")

	defer func() {
		err = outFile.Close()
		require.NoError(t, err, "Closing output file")
		err = os.Remove(conf.Flags.GFD.OutputFile)
		require.NoError(t, err, "Removing output file")
	}()

	output, err := ioutil.ReadAll(outFile)
	require.NoError(t, err, "Reading output file")

	err = checkResult(output, cfg.Path("tests/expected-output-mig-single.txt"), false)
	require.NoError(t, err, "Checking result")

	labels, err := buildLabelMapFromOutput(output)
	require.NoError(t, err, "Building map of labels from output file")

	require.Equal(t, labels["nvidia.com/mig.strategy"], "single", "Incorrect label")
	require.Equal(t, labels["nvidia.com/gpu.count"], "1", "Incorrect label")
	require.Equal(t, labels["nvidia.com/gpu.product"], "MOCKMODEL", "Incorrect label")
	require.Equal(t, labels["nvidia.com/gpu.memory"], "128", "Incorrect label")
}

func TestMigStrategySingleForMigDeviceMigDisabled(t *testing.T) {
	nvmlMock := NewTestNvmlMock()
	// create VGPU mock library with empty vgpu devices
	vgpuMock := NewTestVGPUMock()
	nvmlMock.Devices[0].MigEnabled = false
	nvmlMock.Devices[0].MigDevices = []nvml.NvmlMockDevice{
		{
			Model: "MOCKMODEL",
			Attributes: &nvml.DeviceAttributes{
				GpuInstanceSliceCount: 3,
				MemorySizeMB:          20000,
			},
		},
		{
			Model: "MOCKMODEL",
			Attributes: &nvml.DeviceAttributes{
				GpuInstanceSliceCount: 3,
				MemorySizeMB:          20000,
			},
		},
	}

	conf := &config.Config{
		Flags: config.Flags{
			CommandLineFlags: config.CommandLineFlags{
				MigStrategy:     "single",
				FailOnInitError: true,
				GFD: config.GFDCommandLineFlags{
					Oneshot:       true,
					OutputFile:    "./gfd-test-mig-single-no-mig",
					SleepInterval: time.Second,
					NoTimestamp:   false,
				},
			},
		},
	}

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

	outFile, err := os.Open(conf.Flags.GFD.OutputFile)
	require.NoError(t, err, "Opening output file")

	defer func() {
		err = outFile.Close()
		require.NoError(t, err, "Closing output file")
		err = os.Remove(conf.Flags.GFD.OutputFile)
		require.NoError(t, err, "Removing output file")
	}()

	output, err := ioutil.ReadAll(outFile)
	require.NoError(t, err, "Reading output file")

	err = checkResult(output, cfg.Path("tests/expected-output-mig-single.txt"), false)
	require.NoError(t, err, "Checking result")

	labels, err := buildLabelMapFromOutput(output)
	require.NoError(t, err, "Building map of labels from output file")

	require.Equal(t, labels["nvidia.com/mig.strategy"], "single", "Incorrect label")
	require.Equal(t, labels["nvidia.com/gpu.count"], "1", "Incorrect label")
	require.Equal(t, labels["nvidia.com/gpu.product"], "MOCKMODEL", "Incorrect label")
	require.Equal(t, labels["nvidia.com/gpu.memory"], "128", "Incorrect label")
}

func TestMigStrategySingle(t *testing.T) {
	nvmlMock := NewTestNvmlMock()
	// create VGPU mock library with empty vgpu devices
	vgpuMock := NewTestVGPUMock()
	nvmlMock.Devices[0].MigEnabled = true
	nvmlMock.Devices[0].MigDevices = []nvml.NvmlMockDevice{
		{
			Model: "MOCKMODEL",
			Attributes: &nvml.DeviceAttributes{
				GpuInstanceSliceCount: 3,
				MemorySizeMB:          20000,
			},
		},
		{
			Model: "MOCKMODEL",
			Attributes: &nvml.DeviceAttributes{
				GpuInstanceSliceCount: 3,
				MemorySizeMB:          20000,
			},
		},
	}

	conf := &config.Config{
		Flags: config.Flags{
			CommandLineFlags: config.CommandLineFlags{
				MigStrategy:     "single",
				FailOnInitError: true,
				GFD: config.GFDCommandLineFlags{
					Oneshot:       true,
					OutputFile:    "./gfd-test-mig-single",
					SleepInterval: time.Second,
					NoTimestamp:   false,
				},
			},
		},
	}

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

	outFile, err := os.Open(conf.Flags.GFD.OutputFile)
	require.NoError(t, err, "Opening output file")

	defer func() {
		err = outFile.Close()
		require.NoError(t, err, "Closing output file")
		err = os.Remove(conf.Flags.GFD.OutputFile)
		require.NoError(t, err, "Removing output file")
	}()

	output, err := ioutil.ReadAll(outFile)
	require.NoError(t, err, "Reading output file")

	err = checkResult(output, cfg.Path("tests/expected-output-mig-single.txt"), false)
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
	vgpuMock := NewTestVGPUMock()

	nvmlMock.Devices[0].MigEnabled = true
	nvmlMock.Devices[0].MigDevices = []nvml.NvmlMockDevice{
		{
			Attributes: &nvml.DeviceAttributes{
				GpuInstanceSliceCount: 3,
				MemorySizeMB:          20000,
			},
		},
		{
			Attributes: &nvml.DeviceAttributes{
				GpuInstanceSliceCount: 1,
				MemorySizeMB:          5000,
			},
		},
	}

	conf := &config.Config{
		Flags: config.Flags{
			CommandLineFlags: config.CommandLineFlags{
				MigStrategy:     "mixed",
				FailOnInitError: true,
				GFD: config.GFDCommandLineFlags{
					Oneshot:       true,
					OutputFile:    "./gfd-test-mig-mixed",
					SleepInterval: time.Second,
					NoTimestamp:   false,
				},
			},
		},
	}

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

	outFile, err := os.Open(conf.Flags.GFD.OutputFile)
	require.NoError(t, err, "Opening output file")

	defer func() {
		err = outFile.Close()
		require.NoError(t, err, "Closing output file")
		err = os.Remove(conf.Flags.GFD.OutputFile)
		require.NoError(t, err, "Removing output file")
	}()

	output, err := ioutil.ReadAll(outFile)
	require.NoError(t, err, "Reading output file")

	err = checkResult(output, cfg.Path("tests/expected-output-mig-mixed.txt"), false)
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
