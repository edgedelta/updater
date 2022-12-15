#!/bin/bash

set -e

show_usage() {
    echo "Usage: ./$(basename $0) <version> <registry URI>"
}

version=$1
registry=$2

if [[ -z "$version" || -z "$registry" ]]; then
    show_usage
    exit 1
fi

GIT_ROOT=$(git rev-parse --show-toplevel)

KO_DOCKER_REPO=$registry ko build --platform=all --tags $version,latest -B $GIT_ROOT/cmd/agent-updater