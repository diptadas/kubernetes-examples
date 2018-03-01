#!/bin/bash

set -x

GOPATH=$(go env GOPATH)
PACKAGE_NAME=k8s-admission-webhook
REPO_ROOT="$GOPATH/src/$PACKAGE_NAME"
DOCKER_REPO_ROOT="/go/src/$PACKAGE_NAME"
DOCKER_CODEGEN_PKG="/go/src/k8s.io/code-generator"

pushd $REPO_ROOT

rm -rf "$REPO_ROOT"/apis/foocontroller/v1alpha1/*.generated.go

docker run --rm -ti -u $(id -u):$(id -g) \
  -v "$REPO_ROOT":"$DOCKER_REPO_ROOT" \
  -w "$DOCKER_REPO_ROOT" \
  appscode/gengo:release-1.9 "$DOCKER_CODEGEN_PKG"/generate-internal-groups.sh all \
  k8s-admission-webhook/client \
  k8s-admission-webhook/apis \
  k8s-admission-webhook/apis \
  foocontroller:v1alpha1 \
  --go-header-file "$DOCKER_REPO_ROOT/hack/boilerplate.go.txt"

popd
