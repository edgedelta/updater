#!/bin/bash

set -e

show_usage() {
    echo "Usage: ./$(basename $0) <image URI> <API URL> <latest tag endpoint>"
}

image_uri=$1
api_url=$2
latest_tag_endpoint=$3

if [[ -z "$image_uri" || -z "$api_url" || -z "$latest_tag_endpoint" ]]; then
    show_usage
    exit 1
fi

GIT_ROOT=$(git rev-parse --show-toplevel)

cronjob_yml=$(mktemp)

cat $GIT_ROOT/deploy/cronjob.yml.tmpl \
| sed "s|{IMAGE_URI}|$image_uri|g" \
| sed "s|{ED_API_URL}|$api_url|g" \
| sed "s|{ED_LATEST_TAG_ENDPOINT}|$latest_tag_endpoint|g" \
> $cronjob_yml

kubectl delete --ignore-not-found=true -f $GIT_ROOT/deploy/rbac.yml
kubectl delete --ignore-not-found=true -f $GIT_ROOT/deploy/rolebinding.yml
kubectl delete --ignore-not-found=true -f $cronjob_yml

kubectl apply -f $GIT_ROOT/deploy/rbac.yml
kubectl apply -f $GIT_ROOT/deploy/rolebinding.yml
kubectl apply -f $cronjob_yml

cat $cronjob_yml