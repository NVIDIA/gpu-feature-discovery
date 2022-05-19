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
	"io/ioutil"
	"strings"

	"github.com/NVIDIA/gpu-feature-discovery/internal/nvml"
)

type common struct {
	nvml            nvml.Nvml
	machineTypePath string
}

// NewCommonLabeler creates a labeler for generating common NVML-based labels
func NewCommonLabeler(nvml nvml.Nvml, machineTypePath string) Labeler {
	l := common{
		nvml:            nvml,
		machineTypePath: machineTypePath,
	}

	return l
}

// Labels generates common (non-MIG) NVML-based labels
// TODO: We should call nvml.Init here and also return an empty list if no devices
// are present.
func (labeler common) Labels() (Labels, error) {
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

	device, err := labeler.nvml.NewDevice(0)
	if err != nil {
		return nil, fmt.Errorf("error getting device: %v", err)
	}

	computeMajor, computeMinor, err := device.GetCudaComputeCapability()
	if err != nil {
		return nil, fmt.Errorf("failed to determine CUDA compute capability: %v", err)
	}

	labels, err := labeler.getMachineTypeLabels()
	if err != nil {
		return nil, fmt.Errorf("falied to generate machine type label: %v", err)
	}

	labels["nvidia.com/cuda.driver.major"] = driverMajor
	labels["nvidia.com/cuda.driver.minor"] = driverMinor
	labels["nvidia.com/cuda.driver.rev"] = driverRev
	labels["nvidia.com/cuda.runtime.major"] = fmt.Sprintf("%d", *cudaMajor)
	labels["nvidia.com/cuda.runtime.minor"] = fmt.Sprintf("%d", *cudaMinor)
	if computeMajor != 0 {
		family, _ := device.GetArchFamily()
		labels["nvidia.com/gpu.family"] = family
		labels["nvidia.com/gpu.compute.major"] = fmt.Sprintf("%d", computeMajor)
		labels["nvidia.com/gpu.compute.minor"] = fmt.Sprintf("%d", computeMinor)
	}

	return labels, nil
}

func (manager common) getMachineTypeLabels() (Labels, error) {
	data, err := ioutil.ReadFile(manager.machineTypePath)
	if err != nil {
		return nil, fmt.Errorf("error getting machine type: %v", err)
	}
	machineType := strings.TrimSpace(string(data))

	labels := make(map[string]string)
	labels["nvidia.com/gpu.machine"] = strings.Replace(machineType, " ", "-", -1)

	return labels, nil
}
