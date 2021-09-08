#!/bin/bash

set -x

GOPATH=$(go env GOPATH)
PACKAGE_NAME=github.com/diptadas/k8s-extension-apiserver
REPO_ROOT="$GOPATH/src/$PACKAGE_NAME"
DOCKER_REPO_ROOT="/go/src/$PACKAGE_NAME"
DOCKER_CODEGEN_PKG="/go/src/k8s.io/code-generator"

pushd $REPO_ROOT

rm -rf "$REPO_ROOT"/apis/foocontroller/v1alpha1/*.generated.go

docker run --rm -ti -u $(id -u):$(id -g) \
  -v "$REPO_ROOT":"$DOCKER_REPO_ROOT" \
  -w "$DOCKER_REPO_ROOT" \
  appscode/gengo:release-1.9 "$DOCKER_CODEGEN_PKG"/generate-internal-groups.sh deepcopy \
  "$PACKAGE_NAME"/client \
  "$PACKAGE_NAME"/apis \
  "$PACKAGE_NAME"/apis \
  foocontroller:v1alpha1 \
  --go-header-file "$DOCKER_REPO_ROOT/hack/boilerplate.go.txt"

popd
