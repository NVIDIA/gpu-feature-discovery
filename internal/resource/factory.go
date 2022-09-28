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

package resource

import (
	spec "github.com/NVIDIA/k8s-device-plugin/api/config/v1"
)

// NewManager is a factory method that creates a resource Manager based on the specified config.
func NewManager(config *spec.Config) (Manager, error) {
	return WithConfig(getManager(), config), nil
}

// WithConfig modifies a manager depending on the specified config.
// If failure on a call to init is allowed, the manager is wrapped to allow fallback to a Null manager.
func WithConfig(manager Manager, config *spec.Config) Manager {
	return manager
}

// getManager returns the resource manager depending on the system configuration.
func getManager() Manager {
	return NewNVMLManager()
}
