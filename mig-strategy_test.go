// Copyright (c) 2021, NVIDIA CORPORATION. All rights reserved.

package main

import (
	"testing"

	"github.com/NVIDIA/gpu-monitoring-tools/bindings/go/nvml"
	"github.com/stretchr/testify/require"
)

func TestNoneStrategyReturnsLabelsForSingleDevice(t *testing.T) {
	nvmlMock := NewTestNvmlMock()

	none, _ := NewMigStrategy(MigStrategyNone, nvmlMock)
	labels, err := none.GenerateLabels()

	require.NoError(t, err)
	require.Equal(t, "MOCKMODEL", labels["nvidia.com/gpu.product"], "Incorrect label nvidia.com/gpu.product")
	require.Equal(t, "128", labels["nvidia.com/gpu.memory"], "Incorrect label nvidia.com/gpu.memory")
	require.Equal(t, "1", labels["nvidia.com/gpu.count"], "Incorrect label nvidia.com/gpu.count")
	require.Len(t, labels, 3)
}

func TestNoneStrategyReturnsLabelsForTwoIdenticalDevices(t *testing.T) {
	nvmlMock := NewTestNvmlMock()
	nvmlMock.devices = append(nvmlMock.devices, nvmlMock.devices[0])

	model0 := "MOCKMODEL128"
	mem0 := uint64(128)
	model1 := "MOCKMODEL128"
	mem1 := uint64(128)

	nvmlMock.devices = []NvmlMockDevice{
		NvmlMockDevice{
			instance:   &nvml.Device{Model: &model0, Memory: &mem0},
			attributes: &nvml.DeviceAttributes{},
		},
		NvmlMockDevice{
			instance:   &nvml.Device{Model: &model1, Memory: &mem1},
			attributes: &nvml.DeviceAttributes{},
		},
	}

	none, _ := NewMigStrategy(MigStrategyNone, nvmlMock)
	labels, err := none.GenerateLabels()

	require.NoError(t, err)
	require.Equal(t, "MOCKMODEL128", labels["nvidia.com/gpu.product"], "Incorrect label nvidia.com/gpu.product")
	require.Equal(t, "128", labels["nvidia.com/gpu.memory"], "Incorrect label nvidia.com/gpu.memory")
	require.Equal(t, "2", labels["nvidia.com/gpu.count"], "Incorrect label nvidia.com/gpu.count")
	require.Len(t, labels, 3)
}

func TestNoneStrategyReturnsOnlyCountLabelForTwoNonIdenticalDevices(t *testing.T) {
	nvmlMock := NewTestNvmlMock()
	nvmlMock.devices = append(nvmlMock.devices, nvmlMock.devices[0])

	model0 := "MOCKMODEL128"
	mem0 := uint64(128)
	model1 := "MOCKMODEL256"
	mem1 := uint64(256)

	nvmlMock.devices = []NvmlMockDevice{
		NvmlMockDevice{
			instance:   &nvml.Device{Model: &model0, Memory: &mem0},
			attributes: &nvml.DeviceAttributes{},
		},
		NvmlMockDevice{
			instance:   &nvml.Device{Model: &model1, Memory: &mem1},
			attributes: &nvml.DeviceAttributes{},
		},
	}

	none, _ := NewMigStrategy(MigStrategyNone, nvmlMock)
	labels, err := none.GenerateLabels()

	require.NoError(t, err)
	require.Equal(t, "2", labels["nvidia.com/gpu.count"], "Incorrect label nvidia.com/gpu.count")
	require.Len(t, labels, 1)
}

func TestNoneStrategyReturnsWorkingDeviceWhenOneModelNil(t *testing.T) {
	nvmlMock := NewTestNvmlMock()
	nvmlMock.devices = append(nvmlMock.devices, nvmlMock.devices[0])

	model1 := "MOCKMODEL256"
	mem1   := uint64(256)

	nvmlMock.devices = []NvmlMockDevice{
		NvmlMockDevice{
			instance: &nvml.Device{Model: nil, Memory: nil},
			attributes: &nvml.DeviceAttributes{},
		},
		NvmlMockDevice{
			instance: &nvml.Device{Model: &model1, Memory: &mem1},
			attributes: &nvml.DeviceAttributes{},
		},
	}

	none, _ := NewMigStrategy(MigStrategyNone, nvmlMock)
	labels, err := none.GenerateLabels()

	require.NoError(t, err)
	require.Equal(t, "MOCKMODEL256", labels["nvidia.com/gpu.product"], "Incorrect label nvidia.com/gpu.product")
	require.Equal(t, "256", labels["nvidia.com/gpu.memory"], "Incorrect label nvidia.com/gpu.memory")
	require.Equal(t, "2", labels["nvidia.com/gpu.count"], "Incorrect label nvidia.com/gpu.count")
	require.Len(t, labels, 3)
}

func TestSingleStrategyReturnsNoneForSingleDeviceMigDisabled(t *testing.T) {
	nvmlMock := NewTestNvmlMock()

	single, _ := NewMigStrategy(MigStrategySingle, nvmlMock)
	labels, err := single.GenerateLabels()

	require.NoError(t, err)

	require.Equal(t, labels["nvidia.com/mig.strategy"], "single", "Incorrect label")
	require.Len(t, labels, 4)
}

func TestSingleStrategyReturnsNoneForMultipleDevicesMigDisabled(t *testing.T) {
	nvmlMock := NewTestNvmlMock()
	nvmlMock.devices = append(nvmlMock.devices, nvmlMock.devices[0])

	nvmlMock.devices[0].migEnabled = false
	nvmlMock.devices[1].migEnabled = false

	single, _ := NewMigStrategy(MigStrategySingle, nvmlMock)
	labels, err := single.GenerateLabels()

	require.NoError(t, err)

	require.Equal(t, labels["nvidia.com/mig.strategy"], "single", "Incorrect label")
	require.Len(t, labels, 4)
}

func TestSingleStrategyReturnsErrorMixedMigEnabled(t *testing.T) {
	nvmlMock := NewTestNvmlMock()
	nvmlMock.devices = append(nvmlMock.devices, nvmlMock.devices[0])

	nvmlMock.devices[0].migEnabled = true
	nvmlMock.devices[1].migEnabled = false

	single, _ := NewMigStrategy(MigStrategySingle, nvmlMock)
	labels, err := single.GenerateLabels()

	require.Error(t, err)
	require.Nil(t, labels)
}

func TestSingleStrategyReturnsErrorMigEnabledNoMigs(t *testing.T) {
	nvmlMock := NewTestNvmlMock()
	nvmlMock.devices = append(nvmlMock.devices, nvmlMock.devices[0])

	nvmlMock.devices[0].migEnabled = true
	nvmlMock.devices[0].migDevices = []NvmlMockDevice{}
	nvmlMock.devices[1].migEnabled = true

	single, _ := NewMigStrategy(MigStrategySingle, nvmlMock)
	labels, err := single.GenerateLabels()

	require.Error(t, err)
	require.Nil(t, labels)
}

func TestSingleStrategyReturnsErrorMigEnabledMismatchedSlices(t *testing.T) {
	nvmlMock := NewTestNvmlMock()
	nvmlMock.devices = append(nvmlMock.devices, nvmlMock.devices[0])

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
	}
	nvmlMock.devices[1].migEnabled = true
	nvmlMock.devices[1].migDevices = []NvmlMockDevice{
		NvmlMockDevice{
			instance: &nvml.Device{
				Model: nvmlMock.devices[0].Instance().Model,
			},
			attributes: &nvml.DeviceAttributes{
				GpuInstanceSliceCount: 5,
				MemorySizeMB:          20000,
			},
		},
	}

	single, _ := NewMigStrategy(MigStrategySingle, nvmlMock)
	labels, err := single.GenerateLabels()

	require.Error(t, err)
	require.Nil(t, labels)
}

func TestSingleStrategyReturnsLabelsMigEnabledMatchedSlices(t *testing.T) {
	nvmlMock := NewTestNvmlMock()
	nvmlMock.devices = append(nvmlMock.devices, nvmlMock.devices[0])

	nvmlMock.devices[0].migEnabled = true
	nvmlMock.devices[0].migDevices = []NvmlMockDevice{
		NvmlMockDevice{
			instance: &nvml.Device{
				Model: nvmlMock.devices[0].Instance().Model,
			},
			attributes: &nvml.DeviceAttributes{
				GpuInstanceSliceCount: 3,
				MemorySizeMB:          20096,
			},
		},
	}
	nvmlMock.devices[1].migEnabled = true
	nvmlMock.devices[1].migDevices = []NvmlMockDevice{
		NvmlMockDevice{
			instance: &nvml.Device{
				Model: nvmlMock.devices[0].Instance().Model,
			},
			attributes: &nvml.DeviceAttributes{
				GpuInstanceSliceCount: 3,
				MemorySizeMB:          20096,
			},
		},
	}

	single, _ := NewMigStrategy(MigStrategySingle, nvmlMock)
	labels, err := single.GenerateLabels()

	require.NoError(t, err)
	require.NotNil(t, labels)

	require.Equal(t, labels["nvidia.com/mig.strategy"], "single", "Incorrect label")

	require.Equal(t, labels["nvidia.com/gpu.count"], "2", "Incorrect label")
	require.Equal(t, labels["nvidia.com/gpu.product"], "MOCKMODEL-MIG-3g.20gb", "Incorrect label")
	require.Equal(t, labels["nvidia.com/gpu.memory"], "20096", "Incorrect label")
}
