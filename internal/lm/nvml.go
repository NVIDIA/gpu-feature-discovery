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
	"strconv"
	"strings"

	"github.com/NVIDIA/gpu-feature-discovery/internal/resource"
	spec "github.com/NVIDIA/k8s-device-plugin/api/config/v1"
)

type nvmlLabeler struct {
	manager  resource.Manager
	config   *spec.Config
	labelers list
}

type cudaLabeler struct {
	manager resource.Manager
}

type migCapabilityLabeler struct {
	manager resource.Manager
}

// NewNVMLLabeler creates a new NVML-based labeler using the provided NVML library and config.
func NewNVMLLabeler(manager resource.Manager, config *spec.Config, machineTypePath string) (Labeler, error) {
	if err := manager.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize NVML: %v", err)
	}
	defer manager.Shutdown()

	machineTypeLabeler, err := newMachineTypeLabeler(machineTypePath)
	if err != nil {
		return nil, fmt.Errorf("failed to construct machine type labeler: %v", err)
	}

	cudaLabeler := cudaLabeler{
		manager: manager,
	}

	migCapabilityLabeler, err := NewMigCapabilityLabeler(manager)
	if err != nil {
		return nil, fmt.Errorf("error creating mig capability labeler: %v", err)
	}

	resourceLabeler, err := NewResourceLabeler(manager, config)
	if err != nil {
		return nil, fmt.Errorf("error creating resource labeler: %v", err)
	}

	l := nvmlLabeler{
		manager: manager,
		config:  config,
		labelers: list{
			machineTypeLabeler,
			cudaLabeler,
			migCapabilityLabeler,
			resourceLabeler,
		},
	}

	return l, nil
}

// Labels generates NVML-based labels
func (labeler nvmlLabeler) Labels() (Labels, error) {
	if err := labeler.manager.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize NVML: %v", err)
	}
	defer func() {
		err := labeler.manager.Shutdown()
		if err != nil {
			fmt.Printf("Warning: Shutdown of NVML returned: %v", err)
		}
	}()

	count, err := labeler.manager.GetDeviceCount()
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
	driverVersion, err := labeler.manager.GetDriverVersion()
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

	cudaMajor, cudaMinor, err := labeler.manager.GetCudaDriverVersion()
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

// NewMigCapabilityLabeler creates a new MIG capability labeler using the provided NVML library
func NewMigCapabilityLabeler(manager resource.Manager) (Labeler, error) {
	l := migCapabilityLabeler{
		manager: manager,
	}
	return l, nil
}

// Labels generates MIG capability label by checking all GPUs on the node
func (labeler migCapabilityLabeler) Labels() (Labels, error) {
	isMigCapable := false
	n, err := labeler.manager.GetDeviceCount()
	if err != nil {
		return nil, err
	}
	if n == 0 {
		// no devices, return empty labels
		return nil, nil
	}

	// loop through all devices to check if any one of them is MIG capable
	for i := 0; i < n; i++ {
		d, err := labeler.manager.GetDeviceByIndex(i)
		if err != nil {
			return nil, err
		}

		isMigCapable, err = d.IsMigCapable()
		if err != nil {
			return nil, fmt.Errorf("error getting mig capability: %v", err)
		}
		if isMigCapable {
			break
		}
	}

	labels := Labels{
		"nvidia.com/mig.capable": strconv.FormatBool(isMigCapable),
	}
	return labels, nil
}
