/*
 * Copyright (c) 2020, NVIDIA CORPORATION.  All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

// Constants representing different MIG strategies.
const (
	MigStrategyNone = "none"
)

// MigStrategy defines the strategy to use for setting labels on MIG devices.
type MigStrategy interface {
	GenerateLabels() ([]byte, error)
}

// NewMigStrategy creates a new MIG strategy to generate labels with.
func NewMigStrategy(strategy string, machineTypePath string, nvml Nvml) (MigStrategy, error) {
	switch strategy {
	case MigStrategyNone:
		return &migStrategyNone{machineTypePath, nvml}, nil
	}
	return nil, fmt.Errorf("Unknown strategy: %v", strategy)
}

type migStrategyNone struct {
	machineTypePath string
	nvml            Nvml
}

func getArchFamily(computeMajor, computeMinor int) string {
	switch computeMajor {
	case 1:
		return "tesla"
	case 2:
		return "fermi"
	case 3:
		return "kepler"
	case 5:
		return "maxwell"
	case 6:
		return "pascal"
	case 7:
		if computeMinor < 5 {
			return "volta"
		}
		return "turing"
	case 8:
		return "ampere"
	}
	return "undefined"
}

func getMachineType(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}

// migStrategyNone
func (s *migStrategyNone) GenerateLabels() ([]byte, error) {
	count, err := s.nvml.GetDeviceCount()
	if err != nil {
		return nil, fmt.Errorf("Error getting device count: %v", err)
	}

	device, err := s.nvml.NewDevice(0)
	if err != nil {
		return nil, fmt.Errorf("Error getting device: %v", err)
	}

	driverVersion, err := s.nvml.GetDriverVersion()
	if err != nil {
		return nil, fmt.Errorf("Error getting driver version: %v", err)
	}

	driverVersionSplit := strings.Split(driverVersion, ".")
	if len(driverVersionSplit) > 3 || len(driverVersionSplit) < 2 {
		return nil, fmt.Errorf("Error getting driver version: Version \"%s\" does not match format \"X.Y[.Z]\"", driverVersion)
	}

	driverMajor := driverVersionSplit[0]
	driverMinor := driverVersionSplit[1]
	driverRev := ""
	if len(driverVersionSplit) > 2 {
		driverRev = driverVersionSplit[2]
	}

	cudaMajor, cudaMinor, err := s.nvml.GetCudaDriverVersion()
	if err != nil {
		return nil, fmt.Errorf("Error getting cuda driver version: %v", err)
	}

	machineType, err := getMachineType(s.machineTypePath)
	if err != nil {
		return nil, fmt.Errorf("Error getting machine type: %v", err)
	}

	output := new(bytes.Buffer)

	fmt.Fprintf(output, "nvidia.com/gfd.timestamp=%d\n", time.Now().Unix())
	fmt.Fprintf(output, "nvidia.com/cuda.driver.major=%s\n", driverMajor)
	fmt.Fprintf(output, "nvidia.com/cuda.driver.minor=%s\n", driverMinor)
	fmt.Fprintf(output, "nvidia.com/cuda.driver.rev=%s\n", driverRev)
	fmt.Fprintf(output, "nvidia.com/cuda.runtime.major=%d\n", *cudaMajor)
	fmt.Fprintf(output, "nvidia.com/cuda.runtime.minor=%d\n", *cudaMinor)
	fmt.Fprintf(output, "nvidia.com/gpu.machine=%s\n", strings.Replace(machineType, " ", "-", -1))
	fmt.Fprintf(output, "nvidia.com/gpu.count=%d\n", count)
	if device.Instance().Model != nil {
		model := strings.Replace(*device.Instance().Model, " ", "-", -1)
		fmt.Fprintf(output, "nvidia.com/gpu.product=%s\n", model)
	}
	if device.Instance().Memory != nil {
		memory := *device.Instance().Memory
		fmt.Fprintf(output, "nvidia.com/gpu.memory=%d\n", memory)
	}
	if device.Instance().CudaComputeCapability.Major != nil {
		major := *device.Instance().CudaComputeCapability.Major
		minor := *device.Instance().CudaComputeCapability.Minor
		family := getArchFamily(major, minor)
		fmt.Fprintf(output, "nvidia.com/gpu.family=%s\n", family)
		fmt.Fprintf(output, "nvidia.com/gpu.compute.major=%d\n", major)
		fmt.Fprintf(output, "nvidia.com/gpu.compute.minor=%d\n", minor)
	}

	return output.Bytes(), nil
}
