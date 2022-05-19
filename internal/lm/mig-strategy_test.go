/**
# Copyright (c) 2021-2022, NVIDIA CORPORATION.  All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
**/

package lm

import (
	"testing"

	"github.com/NVIDIA/gpu-feature-discovery/internal/nvml"
	"github.com/stretchr/testify/require"
)

func TestMigStrategySingleLabels(t *testing.T) {
	mockModel := "MOCKMODEL"
	mockMemory := uint64(300)
	mockMigMemory := uint64(100)

	testCases := []struct {
		description    string
		devices        []nvml.MockDevice
		expectedError  bool
		expectedLabels map[string]string
	}{
		{
			description:   "no devices raises error",
			expectedError: true,
		},
		{
			description: "single non-mig device returns non-mig (none) labels",
			devices: []nvml.MockDevice{
				{
					Model:       "MOCKMODEL",
					TotalMemory: mockMemory,
					MigEnabled:  false,
				},
			},
			expectedLabels: map[string]string{
				"nvidia.com/gpu.count":    "1",
				"nvidia.com/gpu.memory":   "300",
				"nvidia.com/gpu.product":  mockModel,
				"nvidia.com/mig.strategy": "single",
			},
		},
		{
			description: "multiple non-mig device returns non-mig (none) labels",
			devices: []nvml.MockDevice{
				{
					Model:       "MOCKMODEL",
					TotalMemory: mockMemory,
					MigEnabled:  false,
				},
				{
					Model:       "MOCKMODEL",
					TotalMemory: mockMemory,
					MigEnabled:  false,
				},
			},
			expectedLabels: map[string]string{
				"nvidia.com/gpu.count":    "2",
				"nvidia.com/gpu.memory":   "300",
				"nvidia.com/gpu.product":  mockModel,
				"nvidia.com/mig.strategy": "single",
			},
		},
		{
			description: "single mig-enabled device returns mig labels",
			devices: []nvml.MockDevice{
				{
					Model:       "MOCKMODEL",
					TotalMemory: mockMemory,
					MigEnabled:  true,
					MigDevices: []nvml.MockDevice{
						{
							Attributes: &nvml.DeviceAttributes{
								MemorySizeMB:              mockMigMemory,
								GpuInstanceSliceCount:     1,
								ComputeInstanceSliceCount: 2,
							},
						},
					},
				},
			},
			expectedLabels: map[string]string{
				"nvidia.com/gpu.count":           "1",
				"nvidia.com/gpu.memory":          "100",
				"nvidia.com/gpu.product":         "MOCKMODEL-MIG-1g.1gb",
				"nvidia.com/mig.strategy":        "single",
				"nvidia.com/gpu.multiprocessors": "0",
				"nvidia.com/gpu.slices.gi":       "1",
				"nvidia.com/gpu.slices.ci":       "2",
				"nvidia.com/gpu.engines.copy":    "0",
				"nvidia.com/gpu.engines.decoder": "0",
				"nvidia.com/gpu.engines.encoder": "0",
				"nvidia.com/gpu.engines.jpeg":    "0",
				"nvidia.com/gpu.engines.ofa":     "0",
			},
		},
		{
			description: "multiple mig-enabled devices returns mig labels",
			devices: []nvml.MockDevice{
				{
					Model:       "MOCKMODEL",
					TotalMemory: mockMemory,
					MigEnabled:  true,
					MigDevices: []nvml.MockDevice{
						{
							Attributes: &nvml.DeviceAttributes{
								MemorySizeMB:              mockMigMemory,
								GpuInstanceSliceCount:     1,
								ComputeInstanceSliceCount: 2,
								MultiprocessorCount:       12,
								SharedCopyEngineCount:     13,
								SharedDecoderCount:        14,
								SharedEncoderCount:        15,
								SharedJpegCount:           16,
								SharedOfaCount:            17,
							},
						},
					},
				},
				{
					Model:       "MOCKMODEL",
					TotalMemory: mockMemory,
					MigEnabled:  true,
					MigDevices: []nvml.MockDevice{
						{
							Attributes: &nvml.DeviceAttributes{
								MemorySizeMB:              mockMigMemory,
								GpuInstanceSliceCount:     1,
								ComputeInstanceSliceCount: 2,
								MultiprocessorCount:       12,
								SharedCopyEngineCount:     13,
								SharedDecoderCount:        14,
								SharedEncoderCount:        15,
								SharedJpegCount:           16,
								SharedOfaCount:            17,
							},
						},
					},
				},
			},
			expectedLabels: map[string]string{
				"nvidia.com/gpu.count":           "2",
				"nvidia.com/gpu.memory":          "100",
				"nvidia.com/gpu.product":         "MOCKMODEL-MIG-1g.1gb",
				"nvidia.com/mig.strategy":        "single",
				"nvidia.com/gpu.multiprocessors": "12",
				"nvidia.com/gpu.slices.gi":       "1",
				"nvidia.com/gpu.slices.ci":       "2",
				"nvidia.com/gpu.engines.copy":    "13",
				"nvidia.com/gpu.engines.decoder": "14",
				"nvidia.com/gpu.engines.encoder": "15",
				"nvidia.com/gpu.engines.jpeg":    "16",
				"nvidia.com/gpu.engines.ofa":     "17",
			},
		},
		{
			description: "empty mig devices returns MIG invalid label",
			devices: []nvml.MockDevice{
				{
					Model:       "MOCKMODEL",
					TotalMemory: mockMemory,
					MigEnabled:  true,
				},
			},
			expectedLabels: map[string]string{
				"nvidia.com/gpu.count":    "0",
				"nvidia.com/gpu.memory":   "0",
				"nvidia.com/gpu.product":  "MOCKMODEL-MIG-INVALID",
				"nvidia.com/mig.strategy": "single",
			},
		},
		{
			description: "mixed mig config returns MIG invalid label",
			devices: []nvml.MockDevice{
				{
					Model:       "MOCKMODEL",
					TotalMemory: mockMemory,
					MigEnabled:  true,
					MigDevices: []nvml.MockDevice{
						{
							Attributes: &nvml.DeviceAttributes{
								MemorySizeMB:              mockMigMemory,
								GpuInstanceSliceCount:     1,
								ComputeInstanceSliceCount: 2,
							},
						},
						{
							Attributes: &nvml.DeviceAttributes{
								MemorySizeMB:              mockMigMemory,
								GpuInstanceSliceCount:     3,
								ComputeInstanceSliceCount: 4,
							},
						},
					},
				},
			},
			expectedLabels: map[string]string{
				"nvidia.com/gpu.count":    "0",
				"nvidia.com/gpu.memory":   "0",
				"nvidia.com/gpu.product":  "MOCKMODEL-MIG-INVALID",
				"nvidia.com/mig.strategy": "single",
			},
		},
		{
			description: "mixed mig enabled and disabled returns invalid config",
			devices: []nvml.MockDevice{
				{
					Model:       "MOCKMODEL",
					TotalMemory: mockMemory,
					MigEnabled:  true,
					MigDevices: []nvml.MockDevice{
						{
							Attributes: &nvml.DeviceAttributes{
								MemorySizeMB:              mockMigMemory,
								GpuInstanceSliceCount:     1,
								ComputeInstanceSliceCount: 2,
							},
						},
					},
				},
				{
					Model:       "MOCKMODEL",
					TotalMemory: mockMemory,
					MigEnabled:  false,
				},
			},
			expectedLabels: map[string]string{
				"nvidia.com/gpu.count":    "0",
				"nvidia.com/gpu.memory":   "0",
				"nvidia.com/gpu.product":  "MOCKMODEL-MIG-INVALID",
				"nvidia.com/mig.strategy": "single",
			},
		},
		{
			description: "enabled, disabled, and empty returns invalid config",
			devices: []nvml.MockDevice{
				{
					Model:       "MOCKMODEL",
					TotalMemory: mockMemory,
					MigEnabled:  true,
					MigDevices: []nvml.MockDevice{
						{
							Attributes: &nvml.DeviceAttributes{
								MemorySizeMB:              mockMigMemory,
								GpuInstanceSliceCount:     1,
								ComputeInstanceSliceCount: 2,
							},
						},
					},
				},
				{
					Model:       "MOCKMODEL",
					TotalMemory: mockMemory,
					MigEnabled:  false,
				},
				{
					Model:       "MOCKMODEL",
					TotalMemory: mockMemory,
					MigEnabled:  true,
				},
			},
			expectedLabels: map[string]string{
				"nvidia.com/gpu.count":    "0",
				"nvidia.com/gpu.memory":   "0",
				"nvidia.com/gpu.product":  "MOCKMODEL-MIG-INVALID",
				"nvidia.com/mig.strategy": "single",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			nvmlMock := &nvml.Mock{
				Devices:       tc.devices,
				DriverVersion: "400.300",
				CudaMajor:     1,
				CudaMinor:     1,
			}

			single, _ := NewMigStrategy(MigStrategySingle, nvmlMock)

			labels, err := single.Labels()
			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.EqualValues(t, tc.expectedLabels, labels)
		})
	}
}
