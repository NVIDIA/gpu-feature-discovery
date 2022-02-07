# Copyright (c) 2021, NVIDIA CORPORATION.  All rights reserved.
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

DOCKER   ?= docker
MKDIR    ?= mkdir

##### Global variables #####
include $(CURDIR)/versions.mk

ifeq ($(IMAGE_NAME),)
REGISTRY ?= nvidia
IMAGE_NAME := $(REGISTRY)/gpu-feature-discovery
endif

IMAGE_TAG ?= $(VERSION)-$(DIST)
IMAGE = $(IMAGE_NAME):$(IMAGE_TAG)

##### Public rules #####
DEFAULT_PUSH_TARGET := ubi8
TARGETS := ubi8

BUILD_TARGETS := $(patsubst %,build-%,$(TARGETS))
PUSH_TARGETS := $(patsubst %,push-%,$(TARGETS))
TEST_TARGETS := $(patsubst %,test-%, $(TARGETS))

.PHONY: $(TARGETS) $(PUSH_TARGETS) $(BUILD_TARGETS) $(TEST_TARGETS)

push-%: DIST = $(*)
$(PUSH_TARGETS): push-%:
	$(DOCKER) push "$(IMAGE_NAME):$(IMAGE_TAG)"

# For the default push target we also push a short tag equal to the version.
# We skip this for the development release
DEVEL_RELEASE_IMAGE_VERSION ?= devel
ifneq ($(strip $(VERSION)),$(DEVEL_RELEASE_IMAGE_VERSION))
push-$(DEFAULT_PUSH_TARGET): push-short
endif
push-short:
	$(DOCKER) tag "$(IMAGE_NAME):$(VERSION)-$(DEFAULT_PUSH_TARGET)" "$(IMAGE_NAME):$(VERSION)"
	$(DOCKER) push "$(IMAGE_NAME):$(VERSION)"


build-%: DIST = $(*)
build-%: DOCKERFILE_SUFFIX = $(*)
build-%: DOCKERFILE = $(CURDIR)/deployments/container/Dockerfile.$(DOCKERFILE_SUFFIX)

# Use a generic build target to build the relevant images
$(BUILD_TARGETS): build-%:
	DOCKER_BUILDKIT=1 \
		$(DOCKER) build --pull \
		--platform=linux/amd64 \
		--tag $(IMAGE) \
		--build-arg BASE_DIST="$(DIST)" \
		--build-arg CUDA_VERSION="$(CUDA_VERSION)" \
		--build-arg GOLANG_VERSION="$(GOLANG_VERSION)" \
		--build-arg VERSION="$(VERSION)" \
		-f $(DOCKERFILE) \
		$(CURDIR)


