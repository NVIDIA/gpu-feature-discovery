#!/usr/bin/env python3

import re
import sys
import time
import yaml

from kubernetes import client, config, watch


def get_expected_labels_regexs():
    expected_labels = [
        "feature.node.kubernetes.io/gfd-nvidia-driver-version=[0-9.]+",
        "feature.node.kubernetes.io/gfd-nvidia-model=[A-Za-z]+",
        "feature.node.kubernetes.io/gfd-nvidia-memory=[0-9]*",
        "feature.node.kubernetes.io/gfd-nvidia-timestamp=[0-9]{10}",
    ]

    return [re.compile(label) for label in expected_labels]


def deploy_yaml_file(core_api, apps_api, rbac_api, daemonset_yaml_file):
    with open(daemonset_yaml_file) as f:
        bodies = yaml.safe_load_all(f)
        for body in bodies:
            namespace = body["metadata"].get("namespace", "default")
            if body["kind"] == "DaemonSet":
                apps_api.create_namespaced_daemon_set(namespace, body)
            elif body["kind"] == "ServiceAccount":
                core_api.create_namespaced_service_account(namespace, body)
            elif body["kind"] == "ClusterRole":
                rbac_api.create_cluster_role(body)
            elif body["kind"] == "ClusterRoleBinding":
                rbac_api.create_cluster_role_binding(body)
            else:
                print("Unknown kind {}".format(body["kind"]), file=sys.stderr)
                sys.exit(1)


def check_labels(expected_labels_regexs, labels):
    for label in labels[:]:
        if label.startswith("feature.node.kubernetes.io/") and \
            not label.startswith("feature.node.kubernetes.io/gfd-"):
                labels.remove(label)
                continue
        for label_regex in expected_labels_regexs[:]:
            if label_regex.match(label):
                expected_labels_regexs.remove(label_regex)
                labels.remove(label)
                break

    for label in labels:
        print("Unexpected label on node: {}".format(label), file=sys.stderr)

    for regex in expected_labels_regexs:
        print("Missing label matching regex: {}".format(regex.pattern), file=sys.stderr)

    return len(expected_labels_regexs) == 0 and len(labels) == 0


if __name__ == '__main__':

    if len(sys.argv) != 3:
        print("Usage: {} GFD_YAML_PATH NFD_YAML_PATH".format(sys.argv[0]))
        sys.exit(1)

    print("Running E2E tests for GFD")

    config.load_kube_config()
    core_api = client.CoreV1Api()
    apps_api = client.AppsV1Api()
    rbac_api = client.RbacAuthorizationV1Api()

    nodes = core_api.list_node().items

    # Should we limit to only one node ?
    if len(nodes) < 1:
        print("No nodes found", file=sys.stderr)
        sys.exit(1)

    regexs = get_expected_labels_regexs()
    for k, v in nodes[0].metadata.labels.items():
        regexs.append(re.compile(k + "=" + v))

    print("Deploy NFD and GFD")
    # TODO: Use real yamls
    deploy_yaml_file(core_api, apps_api, rbac_api, sys.argv[1]) # GFD
    deploy_yaml_file(core_api, apps_api, rbac_api, sys.argv[2]) # NFD

    timestamp_label_name = "feature.node.kubernetes.io/gfd-nvidia-timestamp"

    print("Watching node updates")
    stop = False
    w = watch.Watch()
    for event in w.stream(core_api.list_node, _request_timeout=180):
        if event['type'] == 'MODIFIED':
            print("Node modified")
            for label_name in event['object'].metadata.labels:
                if label_name == timestamp_label_name:
                    stop = True
                    print("Timestamp label found. Stop watching node")
                    break
        if stop:
            break

    print("Checking labels")
    nodes = core_api.list_node().items
    labels = [k + "=" + v for k, v in nodes[0].metadata.labels.items()]

    if not check_labels(regexs, labels):
        print("E2E tests failed", file=sys.stderr)
        sys.exit(1)

    print("E2E tests done")
    sys.exit(0)
