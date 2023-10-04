image: docker
  services:
    - name: docker:dind
      command: [build.js]

variables:
  GIT_SUBMODULE_STRATEGY: recursive
  AWS_KUBE_CI_PARAMS: "--driver --k8s-plugin"
  TF_VAR_FILE: "${PROJECT_DIR/tests/terraform"
  BUILDIMAGE: "${REGISTRY_IMAGE}/build:${COMMIT_SHORT_SHA}"
  BUILD_MULTI_ARCH_IMAGES: "true"

stages:
  - trigger
  - image
  - lint
  - go-checks
  - go-build
  - unit-tests
  - package-build
  - image-build
  - test
  - scan
  - aws_kube_setup
  - integration_tests
  - e2e_tests
  - aws_kube_clean
  - release

workflow:
  rules:
    - if: $CI_PIPELINE_SOURCE == 'merge_request_event'
    - if: $CI_PIPELINE_SOURCE == "web"
    - if: $CI_COMMIT_BRANCH == "main"
    - if: $CI_COMMIT_BRANCH =~ /^release-.*$/
    - if: $CI_COMMIT_TAG && $CI_COMMIT_TAG != ""

.main-or-manual:
  rules:
    - if: $CI_COMMIT_BRANCH == "main"
    - if: $CI_COMMIT_BRANCH =~ /^release-.*$/
    - if: $CI_COMMIT_TAG && $CI_COMMIT_TAG != ""
    - if: $CI_PIPELINE_SOURCE == "schedule"
      when: manual

trigger-pipeline:
  stage: trigger
  script:
    - echo "starting pipeline"
  rules:
    - !reference [.main-or-manual, rules]
    - if: $CI_PIPELINE_SOURCE == 'merge_request_event'
      when: manual
      allow_failure: false
    - when: always

# Define the distribution targets
.dist-ubi8:
  variables:
    DIST: ubi8
    CVE_UPDATES: "cyrus-sasl-lib"

# Define the platform targets
.platform-amd64:
  variables:
    PLATFORM: linux/amd64

.platform-arm64:
  variables:
    PLATFORM: linux/arm64

# Make buildx available as a docker CLI plugin
.buildx-setup:
  before_script:
    -  export BUILDX_VERSION=v0.6.3
    -  apk add --no-cache curl
    -  mkdir -p ~/.docker/cli-plugins
    -  curl -sSLo ~/.docker/cli-plugins/docker-buildx "https://github.com/docker/buildx/releases/download/${BUILDX_VERSION}/buildx-${BUILDX_VERSION}.linux-amd64"
    -  chmod a+x ~/.docker/cli-plugins/docker-buildx

    -  docker buildx create --use --platform=linux/amd64,linux/arm64

    -  '[[ -n "${SKIP_QEMU_SETUP}" ]] || docker run --rm --privileged multiarch/qemu-user-static --reset -p yes'

# Define test helpers
.integration:
  stage: test
  variables:
    IMAGE_NAME: "${CI_REGISTRY_IMAGE}"
    VERSION: "${CI_COMMIT_SHORT_SHA}"
  before_script:
    - apk add --no-cache make bash jq
    - docker login -u "${CI_REGISTRY_USER}" -p "${CI_REGISTRY_PASSWORD}" "${CI_REGISTRY}"
    - docker pull "${IMAGE_NAME}:${VERSION}-${DIST}"
  script:
    - make -f deployments/container/Makefile test-${DIST}

# Download the regctl binary for use in the release steps
.regctl-setup:
  before_script:
    - export REGCTL_VERSION=v0.3.10
    - apk add --no-cache curl
    - mkdir -p bin
    - curl -sSLo bin/regctl https://github.com/regclient/regclient/releases/download/${REGCTL_VERSION}/regctl-linux-amd64
    - chmod a+x bin/regctl
    - export PATH=$(pwd)/bin:${PATH}

# .release forms the base of the deployment jobs which push images to the CI registry.
# This is extended with the version to be deployed (e.g. the SHA or TAG) and the
# target os.
.release:
  stage: release
  variables:
    # Define the source image for the release
    IMAGE_NAME: "${CI_REGISTRY_IMAGE}"
    VERSION: "${CI_COMMIT_SHORT_SHA}"
    # OUT_IMAGE_VERSION is overridden for external releases
    OUT_IMAGE_VERSION: "${CI_COMMIT_SHORT_SHA}"
  before_script:
    - !reference [.regctl-setup, before_script]

    # We ensure that the OUT_IMAGE_VERSION is set
    - 'echo Version: ${OUT_IMAGE_VERSION} ; [[ -n "${OUT_IMAGE_VERSION}" ]] || exit 1'

    # In the case where we are deploying a different version to the CI_COMMIT_SHA, we
    # need to tag the image.
    # Note: a leading 'v' is stripped from the version if present
    - apk add --no-cache make bash
  script:
    # Log in to the "output" registry, tag the image and push the image
    - 'echo "Logging in to CI registry ${CI_REGISTRY}"'
    - regctl registry login "${CI_REGISTRY}" -u "${CI_REGISTRY_USER}" -p "${CI_REGISTRY_PASSWORD}"
    - '[ ${CI_REGISTRY} = ${OUT_REGISTRY} ] || echo "Logging in to output registry ${OUT_REGISTRY}"'
    - '[ ${CI_REGISTRY} = ${OUT_REGISTRY} ] || regctl registry login "${OUT_REGISTRY}" -u "${OUT_REGISTRY_USER}" -p "${OUT_REGISTRY_TOKEN}"'

    # Since OUT_IMAGE_NAME and OUT_IMAGE_VERSION are set, this will push the CI image to the
    # Target
    - make -f deployments/container/Makefile push-${DIST}

# Define a staging release step that pushes an image to an internal "staging" repository
# This is triggered for all pipelines (i.e. not only tags) to test the pipeline steps
# outside of the release process.
.release:staging:
  extends:
    - .release
  variables:
    OUT_REGISTRY_USER: "${CI_REGISTRY_USER}"
    OUT_REGISTRY_TOKEN: "${CI_REGISTRY_PASSWORD}"
    OUT_REGISTRY: "${CI_REGISTRY}"
    OUT_IMAGE_NAME: "${CI_REGISTRY_IMAGE}/staging/gpu-feature-discovery"

# Define an external release step that pushes an image to an external repository.
# This includes a devlopment image off main.
.release:external:
  extends:
    - .release
  rules:
    - if: $CI_COMMIT_TAG
      variables:
        OUT_IMAGE_VERSION: "${CI_COMMIT_TAG}"
    - if: $CI_COMMIT_BRANCH == $RELEASE_DEVEL_BRANCH
      variables:
        OUT_IMAGE_VERSION: "${DEVEL_RELEASE_IMAGE_VERSION}"

# Define the release jobs
release:staging-ubi8:
  extends:
    - .release:staging
    - .dist-ubi8
  needs:
    - image-ubi8
