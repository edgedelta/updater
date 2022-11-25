#!/bin/bash

set -e

show_usage() {
    echo "Usage: ./$(basename $0) <arch> <registry URI> <ED_MODE>"
}

arch=$1
registry=$2
mode=$3

if [[ -z "$arch" || -z "$registry" || -z "$mode" ]]; then
    show_usage
    exit 1
fi

GIT_ROOT=$(git rev-parse --show-toplevel)

local_tag=$(ko build $GIT_ROOT/cmd/updater --local --platform $arch)

image_tag="$registry:updater-$(echo $arch | sed 's|\/|\-|g')-$mode"

docker tag $local_tag $image_tag > /dev/null
docker push $image_tag > /dev/null

echo $image_tag