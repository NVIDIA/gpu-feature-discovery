package lm

import (
	"testing"

	"github.com/NVIDIA/gpu-feature-discovery/internal/nvml"
	"github.com/stretchr/testify/require"
)

func TestMigCapabilityLabeler(t *testing.T) {
	testCases := []struct {
		description    string
		devices        []nvml.MockDevice
		expectedError  bool
		expectedLabels map[string]string
	}{
		{
			description: "no devices returns empty labels",
		},
		{
			description: "single non-mig capable device returns mig.capable as false",
			devices: []nvml.MockDevice{
				{
					Model:      "MOCKMODEL",
					MigCapable: false,
				},
			},
			expectedLabels: map[string]string{
				"nvidia.com/mig.capable": "false",
			},
		},
		{
			description: "multiple non-mig capable devices returns mig.capable as false",
			devices: []nvml.MockDevice{
				{
					Model:      "MOCKMODEL",
					MigCapable: false,
				},
				{
					Model:      "MOCKMODEL",
					MigCapable: false,
				},
			},
			expectedLabels: map[string]string{
				"nvidia.com/mig.capable": "false",
			},
		},
		{
			description: "single mig capable device returns mig.capable as true",
			devices: []nvml.MockDevice{
				{
					Model:      "MOCKMODEL",
					MigCapable: true,
				},
			},
			expectedLabels: map[string]string{
				"nvidia.com/mig.capable": "true",
			},
		},
		{
			description: "one mig capable device among multiple returns mig.capable as true",
			devices: []nvml.MockDevice{
				{
					Model:      "MOCKMODEL",
					MigCapable: false,
				},
				{
					Model:      "MOCKMODEL",
					MigCapable: true,
				},
			},
			expectedLabels: map[string]string{
				"nvidia.com/mig.capable": "true",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			nvmlMock := &nvml.Mock{
				Devices:       tc.devices,
				DriverVersion: "510.73",
				CudaMajor:     11,
				CudaMinor:     6,
			}

			migCapabilityLabeler, _ := NewMigCapabilityLabeler(nvmlMock)

			labels, err := migCapabilityLabeler.Labels()
			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.EqualValues(t, tc.expectedLabels, labels)
		})
	}
}
