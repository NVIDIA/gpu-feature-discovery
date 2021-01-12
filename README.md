# NVIDIA GPU feature discovery

[![Go Report Card](https://goreportcard.com/badge/github.com/NVIDIA/gpu-feature-discovery)](https://goreportcard.com/report/github.com/NVIDIA/gpu-feature-discovery)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## Table of Contents

- [NVIDIA GPU feature discovery](#nvidia-gpu-feature-discovery)
  * [Overview](#overview)
  * [Beta Version](#beta-version)
  * [Prerequisites](#prerequisites)
  * [Quick Start](#quick-start)
    + [Node Feature Discovery (NFD)](#node-feature-discovery-nfd)
    + [Preparing your GPU Nodes](#preparing-your-gpu-nodes)
    + [Deploy NVIDIA GPU Feature Discovery (GFD)](#deploy-nvidia-gpu-feature-discovery-gfd)
      - [Daemonset](#daemonset)
      - [Job](#job)
    + [Verifying Everything Works](#verifying-everything-works)
  * [The GFD Command line interface](#the-gfd-command-line-interface)
  * [Generated Labels](#generated-labels)
    + [MIG 'single' strategy](#mig-single-strategy)
    + [MIG 'mixed' strategy](#mig-mixed-strategy)
  * [Deployment via `helm`](#deployment-via-helm)
    + [Installing via `helm install`from the `gpu-feature-discovery` `helm` repository](#installing-via-helm-install-from-the-gpu-feature-discovery-helm-repository)
    + [Deploying via `helm install` with a direct URL to the `helm` package](#deploying-via-helm-install-with-a-direct-url-to-the-helm-package)
  * [Building and running locally with Docker](#building-and-running-locally-with-docker)
  * [Building and running locally on your native machine](#building-and-running-locally-on-your-native-machine)

## Overview

NVIDIA GPU Feature Discovery for Kubernetes is a software component that allows
you to automatically generate labels for the set of GPUs available on a node.
It leverages the [Node Feature
Discovery](https://github.com/kubernetes-sigs/node-feature-discovery)
to perform this labeling.

## Beta Version

This tool should be considered beta until it reaches `v1.0.0`. As such, we may
break the API before reaching `v1.0.0`, but we will setup a deprecation policy
to ease the transition.

## Prerequisites

The list of prerequisites for running the NVIDIA GPU Feature Discovery is
described below:
* nvidia-docker version > 2.0 (see how to [install](https://github.com/NVIDIA/nvidia-docker)
and it's [prerequisites](https://github.com/nvidia/nvidia-docker/wiki/Installation-\(version-2.0\)#prerequisites))
* docker configured with nvidia as the [default runtime](https://github.com/NVIDIA/nvidia-docker/wiki/Advanced-topics#default-runtime).
* Kubernetes version >= 1.10
* NVIDIA device plugin for Kubernetes (see how to [setup](https://github.com/NVIDIA/k8s-device-plugin))
* NFD deployed on each node you want to label with the local source configured
  * When deploying GPU feature discovery with helm (as described below) we provide a way to automatically deploy NFD for you
  * To deploy NFD yourself, please see https://github.com/kubernetes-sigs/node-feature-discovery

## Quick Start

The following assumes you have at least one node in your cluster with GPUs and
the standard NVIDIA [drivers](https://www.nvidia.com/Download/index.aspx) have
already been installed on it. 

### Node Feature Discovery (NFD)

The first step is to make sure that [Node Feature Discovery](https://github.com/kubernetes-sigs/node-feature-discovery)
is running on every node you want to label. NVIDIA GPU Feature Discovery use
the `local` source so be sure to mount volumes. See
https://github.com/kubernetes-sigs/node-feature-discovery for more details.

You also need to configure the `Node Feature Discovery` to only expose vendor
IDs in the PCI source. To do so, please refer to the Node Feature Discovery
documentation.

The following command will deploy NFD with the minimum required set of
parameters to run `gpu-feature-discovery`.

```shell
kubectl apply -f https://raw.githubusercontent.com/NVIDIA/gpu-feature-discovery/v0.3.0/deployments/static/nfd.yaml
```

**Note:** This is a simple static daemonset meant to demonstrate the basic
features required of `node-feature-discovery` in order to successfully run
`gpu-feature-discovery`. Please see the instructions below for [Deployment via
`helm`](#deployment-via-helm) when deploying in a production setting.

### Preparing your GPU Nodes

Be sure that [nvidia-docker2](https://github.com/NVIDIA/nvidia-docker) is
installed on your GPU nodes and Docker default runtime is set to `nvidia`. See
https://github.com/NVIDIA/nvidia-docker/wiki/Advanced-topics#default-runtime.

### Deploy NVIDIA GPU Feature Discovery (GFD)

The next step is to run NVIDIA GPU Feature Discovery on each node as a Daemonset
or as a Job.

#### Daemonset

```shell
kubectl apply -f https://raw.githubusercontent.com/NVIDIA/gpu-feature-discovery/v0.3.0/deployments/static/gpu-feature-discovery-daemonset.yaml
```

**Note:** This is a simple static daemonset meant to demonstrate the basic
features required of `gpu-feature-discovery`. Please see the instructions below
for [Deployment via `helm`](#deployment-via-helm) when deploying in a
production setting.

#### Job

You must change the `NODE_NAME` value in the template to match the name of the
node you want to label:

```shell
$ export NODE_NAME=<your-node-name>
$ curl https://raw.githubusercontent.com/NVIDIA/gpu-feature-discovery/v0.3.0/deployments/static/gpu-feature-discovery-job.yaml.template \
    | sed "s/NODE_NAME/${NODE_NAME}/" > gpu-feature-discovery-job.yaml
$ kubectl apply -f gpu-feature-discovery-job.yaml
```

**Note:** This method should only be used for testing and not deployed in a
productions setting.

### Verifying Everything Works

With both NFD and GFD deployed and running, you should now be able to see GPU
related labels appearing on any nodes that have GPUs installed on them.

```
$ kubectl get nodes -o yaml
apiVersion: v1
items:
- apiVersion: v1
  kind: Node
  metadata:
    ...

    labels:
      nvidia.com/cuda.driver.major: "455"
      nvidia.com/cuda.driver.minor: "06"
      nvidia.com/cuda.driver.rev: ""
      nvidia.com/cuda.runtime.major: "11"
      nvidia.com/cuda.runtime.minor: "1"
      nvidia.com/gfd.timestamp: "1594644571"
      nvidia.com/gpu.compute.major: "8"
      nvidia.com/gpu.compute.minor: "0"
      nvidia.com/gpu.count: "1"
      nvidia.com/gpu.family: ampere
      nvidia.com/gpu.machine: NVIDIA DGX-2H
      nvidia.com/gpu.memory: "39538"
      nvidia.com/gpu.product: A100-SXM4-40GB
      ...
...

```

## The GFD Command line interface

Available options:
```
gpu-feature-discovery:
Usage:
  gpu-feature-discovery [--fail-on-init-error=<bool>] [--mig-strategy=<strategy>] [--oneshot | --sleep-interval=<seconds>] [--output-file=<file> | -o <file>]
  gpu-feature-discovery -h | --help
  gpu-feature-discovery --version

Options:
  -h --help                       Show this help message and exit
  --version                       Display version and exit
  --oneshot                       Label once and exit
  --fail-on-init-error=<bool>     Fail if there is an error during initialization of any label sources [Default: true]
  --sleep-interval=<seconds>      Time to sleep between labeling [Default: 60s]
  --mig-strategy=<strategy>       Strategy to use for MIG-related labels [Default: none]
  -o <file> --output-file=<file>  Path to output file
                                  [Default: /etc/kubernetes/node-feature-discovery/features.d/gfd]
```

You can also use environment variables:

| Env Variable           | Option               | Example |
| ---------------------- | -------------------- | ------- |
| GFD_FAIL_ON_INIT_ERROR | --fail-on-init-error | true    |
| GFD_MIG_STRATEGY       | --mig-strategy       | none    |
| GFD_ONESHOT            | --oneshot            | TRUE    |
| GFD_OUTPUT_FILE        | --output-file        | output  |
| GFD_SLEEP_INTERVAL     | --sleep-interval     | 10s     |

Environment variables override the command line options if they conflict.

## Generated Labels

This is the list of the labels generated by NVIDIA GPU Feature Discovery and
their meaning:

| Label Name                     | Value Type | Meaning                                  | Example        |
| -------------------------------| ---------- | ---------------------------------------- | -------------- |
| nvidia.com/cuda.driver.major   | Integer    | Major of the version of NVIDIA driver    | 418            |
| nvidia.com/cuda.driver.minor   | Integer    | Minor of the version of NVIDIA driver    | 30             |
| nvidia.com/cuda.driver.rev     | Integer    | Revision of the version of NVIDIA driver | 40             |
| nvidia.com/cuda.runtime.major  | Integer    | Major of the version of CUDA             | 10             |
| nvidia.com/cuda.runtime.minor  | Integer    | Minor of the version of CUDA             | 1              |
| nvidia.com/gfd.timestamp       | Integer    | Timestamp of the generated labels        | 1555019244     |
| nvidia.com/gpu.compute.major   | Integer    | Major of the compute capabilities        | 3              |
| nvidia.com/gpu.compute.minor   | Integer    | Minor of the compute capabilities        | 3              |
| nvidia.com/gpu.count           | Integer    | Number of GPUs                           | 2              |
| nvidia.com/gpu.family          | String     | Architecture family of the GPU           | kepler         |
| nvidia.com/gpu.machine         | String     | Machine type                             | DGX-1          |
| nvidia.com/gpu.memory          | Integer    | Memory of the GPU in Mb                  | 2048           |
| nvidia.com/gpu.product         | String     | Model of the GPU                         | GeForce-GT-710 |

Depending on the MIG strategy used, the following set of labels may also be
available (or override the default values for some of the labels listed above):

### MIG 'single' strategy

With this strategy, the single `nvidia.com/gpu` label is overloaded to provide
information about MIG devices on the node, rather than full GPUs. This assumes
all GPUs on the node have been divided into identical partitions of the same
size. The example below shows info for a system with 8 full GPUs, each of which
is partitioned into 7 equal sized MIG devices (56 total).

| Label Name                          | Value Type | Meaning                                  | Example                   |
| ----------------------------------- | ---------- | ---------------------------------------- | ------------------------- |
| nvidia.com/mig.strategy             | String     | MIG strategy in use                      | single                    |
| nvidia.com/gpu.product (overridden) | String     | Model of the GPU (with MIG info added)   | A100-SXM4-40GB-MIG-1g.5gb |
| nvidia.com/gpu.count   (overridden) | Integer    | Number of MIG devices                    | 56                        |
| nvidia.com/gpu.memory  (overridden) | Integer    | Memory of each MIG device in Mb          | 5120                      |
| nvidia.com/gpu.multiprocessors      | Integer    | Number of Multiprocessors for MIG device | 14                        |
| nvidia.com/gpu.slices.gi            | Integer    | Number of GPU Instance slices            | 1                         |
| nvidia.com/gpu.slices.ci            | Integer    | Number of Compute Instance slices        | 1                         |
| nvidia.com/gpu.engines.copy         | Integer    | Number of DMA engines for MIG device     | 1                         |
| nvidia.com/gpu.engines.decoder      | Integer    | Number of decoders for MIG device        | 1                         |
| nvidia.com/gpu.engines.encoder      | Integer    | Number of encoders for MIG device        | 1                         |
| nvidia.com/gpu.engines.jpeg         | Integer    | Number of JPEG engines for MIG device    | 0                         |
| nvidia.com/gpu.engines.ofa          | Integer    | Number of OfA engines for MIG device     | 0                         |

### MIG 'mixed' strategy

With this strategy, a separate set of labels for each MIG device type is
generated. The name of each MIG device type is defines as follows:
```
MIG_TYPE=mig-<slice_count>g.<memory_size>.gb
e.g.  MIG_TYPE=mig-3g.20gb
```

| Label Name                           | Value Type | Meaning                                  | Example        |
| ------------------------------------ | ---------- | ---------------------------------------- | -------------- |
| nvidia.com/mig.strategy              | String     | MIG strategy in use                      | mixed          |
| nvidia.com/MIG\_TYPE.count           | Integer    | Number of MIG devices of this type       | 2              |
| nvidia.com/MIG\_TYPE.memory          | Integer    | Memory of MIG device type in Mb          | 10240          |
| nvidia.com/MIG\_TYPE.multiprocessors | Integer    | Number of Multiprocessors for MIG device | 14             |
| nvidia.com/MIG\_TYPE.slices.ci       | Integer    | Number of GPU Instance slices            | 1              |
| nvidia.com/MIG\_TYPE.slices.gi       | Integer    | Number of Compute Instance slices        | 1              |
| nvidia.com/MIG\_TYPE.engines.copy    | Integer    | Number of DMA engines for MIG device     | 1              |
| nvidia.com/MIG\_TYPE.engines.decoder | Integer    | Number of decoders for MIG device        | 1              |
| nvidia.com/MIG\_TYPE.engines.encoder | Integer    | Number of encoders for MIG device        | 1              |
| nvidia.com/MIG\_TYPE.engines.jpeg    | Integer    | Number of JPEG engines for MIG device    | 0              |
| nvidia.com/MIG\_TYPE.engines.ofa     | Integer    | Number of OfA engines for MIG device     | 0              |

## Deployment via `helm`

The preferred method to deploy `gpu-feature-discovery` is as a daemonset using `helm`.
Instructions for installing `helm` can be found
[here](https://helm.sh/docs/intro/install/).

The `helm` chart for the latest release of GFD (`v0.3.0`) includes a number
of customizable values. The most commonly overridden ones are:

```
  failOnInitError:
      Fail if there is an error during initialization of any label sources (default: true)
  sleepInterval:
      time to sleep between labeling (default "60s")
  migStrategy:
      pass the desired strategy for labeling MIG devices on GPUs that support it
      [none | single | mixed] (default "none)
  nfd.deploy:
      When set to true, deploy NFD as a subchart with all of the proper
      parameters set for it (default "true")
      
```

**Note:** The following document provides more information on the available MIG
strategies and how they should be used [Supporting Multi-Instance GPUs (MIG) in
Kubernetes](https://docs.google.com/document/d/1mdgMQ8g7WmaI_XVVRrCvHPFPOMCm5LQD5JefgAh6N8g).

Please take a look in the following `values.yaml` files to see the full set of
overridable parameters for both the top-level `gpu-feature-discovery` chart and
the `node-feature-discovery` subchart.

* https://github.com/NVIDIA/gpu-feature-discovery/blob/v0.3.0/deployments/helm/gpu-feature-discovery/values.yaml
* https://github.com/NVIDIA/gpu-feature-discovery/blob/v0.3.0/deployments/helm/gpu-feature-discovery/charts/node-feature-discovery/values.yaml

#### Installing via `helm install` from the `gpu-feature-discovery` `helm` repository

The preferred method of deployment is with `helm install` via the
`gpu-feature-discovery` `helm` repository.

This repository can be installed as follows:
```shell
$ helm repo add nvgfd https://nvidia.github.io/gpu-feature-discovery
$ helm repo update
```

Once this repo is updated, you can begin installing packages from it to depoloy
the `gpu-feature-discovery` daemonset and (optionally) the
`node-feature-discovery` daemonset. Below are some examples of deploying these
components with the various flags from above.

**Note:** Since this is a pre-release version, you will need to pass the
`--devel` flag to `helm search repo` in order to see this release listed.

Using the default values for all flags:
```shell
$ helm install \
    --version=0.3.0 \
    --generate-name \
    nvgfd/gpu-feature-discovery
```

Disabling auto-deployment of NFD and running with a MIG strategy of 'mixed' in
the default namespace.
```shell
$ helm install \
    --version=0.3.0 \
    --generate-name \
    --set nfd.deploy=false \
    --set migStrategy=mixed
    --set namespace=default \
    nvgfd/gpu-feature-discovery
```

#### Deploying via `helm install` with a direct URL to the `helm` package

If you prefer not to install from the `gpu-feature-discovery` `helm` repo, you can
run `helm install` directly against the tarball of the components `helm` package.
The examples below install the same daemonsets as the method above, except that
they use direct URLs to the `helm` package instead of the `helm` repo.

Using the default values for the flags:
```shell
$ helm install \
    --generate-name \
    https://nvidia.github.com/gpu-feature-discovery/stable/gpu-feature-discovery-0.3.0.tgz
```

Disabling auto-deployment of NFD and running with a MIG strategy of 'mixed' in
the default namespace.
```shell
$ helm install \
    --generate-name \
    --set nfd.deploy=false \
    --set migStrategy=mixed
    --set namespace=default \
    https://nvidia.github.com/gpu-feature-discovery/stable/gpu-feature-discovery-0.3.0.tgz
```

## Building and running locally with Docker

Download the source code:
```shell
git clone https://github.com/NVIDIA/gpu-feature-discovery
```

Build the docker image:
```
export GFD_VERSION=$(git describe --tags --dirty --always)
docker build . --build-arg GFD_VERSION=$GFD_VERSION -t nvidia/gpu-feature-discovery:${GFD_VERSION}
```

Run it:
```
mkdir -p output-dir
docker run -v ${PWD}/output-dir:/etc/kubernetes/node-feature-discovery/features.d nvidia/gpu-feature-discovery:${GFD_VERSION}
```

You should have set the default runtime of Docker to `nvidia` on your host or
you can also use the `--runtime=nvidia` option:
```
docker run --runtime=nvidia nvidia/gpu-feature-discovery:${GFD_VERSION}
```

## Building and running locally on your native machine

Download the source code:
```shell
git clone https://github.com/NVIDIA/gpu-feature-discovery
```

Get dependies:
```shell
dep ensure
```

Build it:
```
export GFD_VERSION=$(git describe --tags --dirty --always)
go build -ldflags "-X main.Version=${GFD_VERSION}"
```

You can also use the Dockerfile.devel:
```
docker build . -f Dockerfile.devel -t gfd-devel
docker run -it gfd-devel
go build -ldflags "-X main.Version=devel"
```

Run it:
```
./gpu-feature-discovery --output=$(pwd)/gfd
```
