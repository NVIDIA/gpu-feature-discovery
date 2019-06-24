#!/bin/sh

set -e

if [ "$#" -lt 2 ]; then
	echo "Usage: $0 IMAGE_NAME VERSION [CI_COMMIT_TAG]" && exit 1
fi

IMAGE_NAME=$1
VERSION=$2
CI_COMMIT_TAG=$3
GFD_YAML_FILE="../gpu-feature-discovery-daemonset.yaml"
NFD_YAML_FILE_URL="https://raw.githubusercontent.com/kubernetes-sigs/node-feature-discovery/v0.4.0/nfd-daemonset-combined.yaml.template"
NFD_YAML_FILE="nfd.yaml"

wget -O "${NFD_YAML_FILE}" "${NFD_YAML_FILE_URL}"

sudo apt install -y python3-pip
sudo pip3 install -r e2e-requirements.txt

# If it's a tag
if [ -n "$CI_COMMIT_TAG" ]; then
	sed -i "s|nvidia/gpu-feature-discovery:|${IMAGE_NAME}:|" ${GFD_YAML_FILE}
else
	sed -i -E "s|nvidia/gpu-feature-discovery:.*|${IMAGE_NAME}:${VERSION}|" ${GFD_YAML_FILE}
fi

./e2e-tests.py ${GFD_YAML_FILE} ${NFD_YAML_FILE}
