#!/bin/bash

set -e

show_usage() {
    echo "Usage: ./$(basename $0) <tags> <registry URI> [platform]"
}

tags=$1
registry=$2
platform=${3:-all}

if [[ -z "$tags" || -z "$registry" ]]; then
    show_usage
    exit 1
fi

GIT_ROOT=$(git rev-parse --show-toplevel)

echo "[+] Tags    : $tags"
echo "[+] Registry: $registry"
echo "[+] Platform: $platform"

if [[ "$registry" -eq "local" ]]; then
    tag=$(ko build --local --platform=$platform --sbom=none --tags $tags -B $GIT_ROOT/cmd/agent-updater)
    docker tag $tag agent-updater:local
else
    KO_DOCKER_REPO=$registry ko build --platform=$platform --sbom=none --tags $tags -B $GIT_ROOT/cmd/agent-updater
fi