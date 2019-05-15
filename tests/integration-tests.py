#!/usr/bin/env python3

import docker
import os
import re
import sys
import tempfile
import time


def get_expected_labels_regexs():
    expected_labels = [		
        "nvidia-driver-version=[0-9.]+",
        "nvidia-model=[A-Za-z]+",
        "nvidia-memory=[0-9]*",
        "nvidia-timestamp=[0-9]{10}",
    ]

    return [re.compile(label) for label in expected_labels]

def check_labels(expected_labels_regexs, labels):
    for label in labels[:]:
        for label_regex in expected_labels_regexs[:]:
            if label_regex.match(label):
                expected_labels_regexs.remove(label_regex)
                labels.remove(label)
                break

    for label in labels:
        print("Unexpected label: {}".format(label))

    for regex in expected_labels_regexs:
        print("Missing label matching regex: {}".format(regex.pattern))

    return len(expected_labels_regexs) == 0 and len(labels) == 0


if __name__ == '__main__':

    if len(sys.argv) != 2:
        print("Usage: {} DOCKER_IMAGE".format(sys.argv[0]))
        sys.exit(1)

    image = sys.argv[1]

    print("Running integration tests for GFD")

    client = docker.from_env()

    with tempfile.TemporaryDirectory() as tmpdirname:
        mount = docker.types.Mount("/etc/kubernetes/node-feature-discovery/features.d",
            tmpdirname, "bind")

        print("Running GFD")

        container = client.containers.run(image, detach=True, mounts=[mount,])

        print("Waiting for GFD output file")

        while not os.path.exists(tmpdirname + "/gfd"):
            time.sleep(1)

        print("GFD output file found. Stoping GFD")

        container.stop()

        with open(tmpdirname + "/gfd") as output_file:
            content = output_file.readlines()
            content = [x.strip() for x in content]
            expected_labels = get_expected_labels_regexs()

            if not check_labels(expected_labels, content):
                print("Integration tests failed")
                sys.exit(1)

            print("Integration tests done")
            sys.exit(0)
