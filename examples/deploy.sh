#!/bin/bash

set -e

kubectl apply -f examples/rbac.yml
kubectl apply -f examples/rolebinding.yml
kubectl apply -f examples/cronjob.yml
