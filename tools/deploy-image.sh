#!/usr/bin/env bash

set -eo pipefail

TARGET_ENVIRONMENT=fire/dev-us-central-0.fire-dev-001
IMAGE_JSONNET_KEY=fire
IMAGE_NAME=us.gcr.io/kubernetes-dev/fire
IMAGE_TAG=main-05800bb

[[ -z "$TARGET_ENVIRONMENT" ]] && echo "TARGET_ENVIRONMENT is required" && exit 1;
[[ -z "$IMAGE_JSONNET_KEY" ]] && echo "IMAGE_JSONNET_KEY is required" && exit 1;
[[ -z "$IMAGE_NAME" ]] && echo "IMAGE_NAME is required" && exit 1;
[[ -z "$IMAGE_TAG" ]] && echo "IMAGE_TAG is required" && exit 1;

docker run \
  -e PLUGIN_GITHUB_TOKEN="${GRAFANABOT_TOKEN}" \
  -e PLUGIN_DOCKER_IMAGE="${IMAGE_NAME}" \
  -e PLUGIN_DOCKER_TAG="${IMAGE_TAG}" \
  -e PLUGIN_JSONNET_KEY="${IMAGE_JSONNET_KEY}" \
  -e PLUGIN_FILE_PATH="ksonnet/environments/${TARGET_ENVIRONMENT}/images.libsonnet" \
  us.gcr.io/kubernetes-dev/drone/plugins/deploy-image@sha256:b4a95200397017e10b771f926988ea8695384d908fad9d1efda4954c417e31c1
