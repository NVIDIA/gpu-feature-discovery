/**
# Copyright (c) 2022, NVIDIA CORPORATION.  All rights reserved.
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
	"fmt"
	"log"
	"strings"

	"github.com/NVIDIA/gpu-feature-discovery/internal/nvml"
	spec "github.com/NVIDIA/k8s-device-plugin/api/config/v1"
)

type nvmlLabeler struct {
	nvml     nvml.Nvml
	config   *spec.Config
	labelers list
}

type cudaLabeler struct {
	nvml nvml.Nvml
}

// NewNVMLLabeler creates a new NVML-based labeler using the provided NVML library and config.
func NewNVMLLabeler(nvml nvml.Nvml, config *spec.Config, machineTypePath string) (Labeler, error) {
	if err := nvml.Init(); err != nil {
		if *config.Flags.FailOnInitError {
			return nil, fmt.Errorf("failed to initialize NVML: %v", err)
		}
		log.Printf("Warning: Error generating NVML labels: %v", err)
		return empty{}, nil
	}

	machineTypeLabeler, err := newMachineTypeLabeler(machineTypePath)
	if err != nil {
		return nil, fmt.Errorf("failed to construct machine type labeler: %v", err)
	}

	cudaLabeler := cudaLabeler{
		nvml: nvml,
	}

	migStrategyLabler, err := NewMigStrategy(*config.Flags.MigStrategy, nvml)
	if err != nil {
		return nil, fmt.Errorf("error creating MIG strategy: %v", err)
	}

	l := nvmlLabeler{
		nvml:   nvml,
		config: config,
		labelers: list{
			machineTypeLabeler,
			cudaLabeler,
			migStrategyLabler,
		},
	}

	return l, nil
}

// Labels generates NVML-based labels
func (labeler nvmlLabeler) Labels() (Labels, error) {
	if err := labeler.nvml.Init(); err != nil {
		if *labeler.config.Flags.FailOnInitError {
			return nil, fmt.Errorf("failed to initialize NVML: %v", err)
		}
		log.Printf("Warning: Error generating NVML labels: %v", err)
		return nil, nil
	}

	defer func() {
		err := labeler.nvml.Shutdown()
		if err != nil {
			fmt.Printf("Warning: Shutdown of NVML returned: %v", err)
		}
	}()

	count, err := labeler.nvml.GetDeviceCount()
	if err != nil {
		return nil, fmt.Errorf("error getting device count: %v", err)
	}

	if count == 0 {
		return nil, nil
	}

	return labeler.labelers.Labels()
}

// Labels generates common (non-MIG) NVML-based labels
func (labeler cudaLabeler) Labels() (Labels, error) {
	driverVersion, err := labeler.nvml.GetDriverVersion()
	if err != nil {
		return nil, fmt.Errorf("error getting driver version: %v", err)
	}

	driverVersionSplit := strings.Split(driverVersion, ".")
	if len(driverVersionSplit) > 3 || len(driverVersionSplit) < 2 {
		return nil, fmt.Errorf("error getting driver version: Version \"%s\" does not match format \"X.Y[.Z]\"", driverVersion)
	}

	driverMajor := driverVersionSplit[0]
	driverMinor := driverVersionSplit[1]
	driverRev := ""
	if len(driverVersionSplit) > 2 {
		driverRev = driverVersionSplit[2]
	}

	cudaMajor, cudaMinor, err := labeler.nvml.GetCudaDriverVersion()
	if err != nil {
		return nil, fmt.Errorf("error getting cuda driver version: %v", err)
	}

	labels := Labels{
		"nvidia.com/cuda.driver.major":  driverMajor,
		"nvidia.com/cuda.driver.minor":  driverMinor,
		"nvidia.com/cuda.driver.rev":    driverRev,
		"nvidia.com/cuda.runtime.major": fmt.Sprintf("%d", *cudaMajor),
		"nvidia.com/cuda.runtime.minor": fmt.Sprintf("%d", *cudaMinor),
	}
	return labels, nil
}
