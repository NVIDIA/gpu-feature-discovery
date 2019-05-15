#!/bin/sh

set -e

if [ "$#" -lt 2 ]; then
	echo "Usage: $0 IMAGE_NAME VERSION [CI_COMMIT_TAG]" && exit 1
fi

IMAGE_NAME=$1
VERSION=$2
CI_COMMIT_TAG=$3
GFD_YAML_FILE=../gpu-feature-discovery-daemonset.yaml
NFD_YAML_FILE=./nfd.yaml

sudo apt install -y python3-pip
sudo pip3 install -r e2e-requirements.txt

# Should be remove once:
# https://github.com/kubernetes-sigs/node-feature-discovery/pull/236 is merged
# and a new version of NFD is released
git clone https://github.com/kubernetes-sigs/node-feature-discovery.git
docker build --build-arg NFD_VERSION=ci -t 127.0.0.1:5000/nfd node-feature-discovery
docker push 127.0.0.1:5000/nfd

# If it's a tag
if [ -n "$CI_COMMIT_TAG" ]; then
	sed -i "s|nvidia/gpu-feature-discovery:|${IMAGE_NAME}:|" ${GFD_YAML_FILE}
else
	sed -i -E "s|nvidia/gpu-feature-discovery:.*|${IMAGE_NAME}:${VERSION}|" ${GFD_YAML_FILE}
fi

./e2e-tests.py ${GFD_YAML_FILE} ${NFD_YAML_FILE}
