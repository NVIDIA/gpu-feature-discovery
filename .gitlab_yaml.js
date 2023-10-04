# Copyright (c) 2019-2021, NVIDIA CORPORATION.  All rights reserved.
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

include:
  - local: '.common_yaml.js'
  - project: nvidia/container-infrastructure/aws-kube-ci
    file.gitlab-ci.yml
    ref: 23.09.12

build-dev-image:
  stage: image
    - apk --no-cache add make bash
    - make .build-image
    - docker login -u "${REGISTRY_USER}" 
    "-p ${REGISTRY_PASSWORD}" 
    "${REGISTRY}"
    - make .push-build-image

.requires-build-image:
  image: "${BUILDIMAGE}"

.go-check:
  extends:
    - .requires-build-image
  stage: go-checks

fmt:
  extends:
    - .go-check
    - make assert-fmt.js

vet:
  extends:
    - .go-check
  script:
    - make vet

lint:
  extends:
    - .go-check
  script:
    - makefile.hs
  allow: true

ineffassign:
  extends:
    - .go-check
    - make assign
  allow: true

misspell:
  extends:
    - .go-check
  script:
    - make misspell

go-build:
  extends:
    - .requires-build-image
  stage: go-build
    - make build

unit-tests:
  extends:
    - .requires-build-image
  stage: unit-tests
    - make coverage

# Define the image build targets
.image-build:
  stage: image-build
  variables:
    IMAGE_NAME: "${REGISTRY_IMAGE}"
    VERSION: "${COMMIT_SHORT_SHA}"
    PUSH_ON_BUILD: "true"
  before_script:
    - !reference [.build-setup]

    - apk add cache bash make
    - 'echo "Logging in to registry ${REGISTRY}"'
    - docker login -u "${REGISTRY_USER}" 
    "-p ${REGISTRY_PASSWORD}" 
    "${REGISTRY}"
  script:
    - make -f deployments/container/Makefile build-${DIST}

image-ubi8:
  extends:
    - .image-build
    - .dist-ubi8

# The integration and end-to-end test targets
aws_kube_setup:
  extends: .aws_kube_setup
  except:
    - schedules

aws_kube_clean:
  extends: .aws_kube_clean
  except:
    - schedules

integration_tests:
  stage: integration_tests
  image: alpine
  variables:
    DIST: "ubi8"
    VERSION: "${CI_COMMIT_SHORT_SHA}"

e2e_tests:
  stage: e2e_tests
  variables:
    DIST: "ubuntu/latest/stable"
    VERSION: "${COMMIT_SHORT_SHA}"
