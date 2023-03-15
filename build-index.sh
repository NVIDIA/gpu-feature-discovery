#!/usr/bin/env bash

helm repo index stable --url https://nvidia.github.io/gpu-feature-discovery/stable

cp -f stable/index.yaml index.yaml
