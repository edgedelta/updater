#!/bin/bash

set -e

arch="linux/arm64"
registry="public.ecr.aws/v4z2v9g0/edgedelta-development"

api_url="https://api.edgedelta.com/v1"
endpoint="/versioning/latest"

if [ "$ED_MODE" = "local-test" ]; then
    # To run the updater with the test API:
    #   `go run test/api/main.go`
    api_url="http://host.minikube.internal:8080"
    endpoint="/"
elif [ "$ED_MODE" = "local" ]; then
    # To run the updater with the admin API locally:
    #   `ED_MODE=staging ED_SECRET_PROVIDER=kms go run cmd/admin/main.go`
    api_url="http://host.minikube.internal:4444/v1"
elif [ "$ED_MODE" = "staging" ]; then
    api_url="https://api.staging.edgedelta.com/v1"
fi

echo "[+] Mode    : $ED_MODE"
echo "[+] Arch    : $arch"
echo "[+] Registry: $registry"
echo "[+] API URL : $api_url"
echo "[+] Endpoint: $endpoint"

GIT_ROOT=$(git rev-parse --show-toplevel)

image_uri=$($GIT_ROOT/deploy/scripts/build.sh $arch $registry $ED_MODE)
$GIT_ROOT/deploy/scripts/deploy.sh $image_uri $api_url $endpoint
