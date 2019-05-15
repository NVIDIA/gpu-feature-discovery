#!/bin/sh

set -e

if [ "$#" -lt 1 ]; then
  echo "Usage: $0 IMAGE" && exit 1
fi

IMAGE=$1

sudo apt install -y python3-pip
sudo pip3 install -r integration-requirements.txt

./integration-tests.py "${IMAGE}"
