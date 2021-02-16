#!/bin/sh

if [ "$#" -lt 1 ]; then
  echo "Usage: $0 VERSION" && exit 1
fi

VERSION=$1
YAML_FILES="../gpu-feature-discovery-daemonset.yaml ../gpu-feature-discovery-job.yaml.template"

ret=0

for file in ${YAML_FILES}; do
  if ! grep -w "nvcr.io/nvidia/gpu-feature-discovery:${VERSION}" "${file}"; then
    echo "GFD image version in YAML ${file} does not match current tag."
    echo "You may have forgotten to update it"
    ret=1
  fi
done
exit $ret
