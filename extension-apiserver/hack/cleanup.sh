#!/usr/bin/env bash
set -x

GOPATH=$(go env GOPATH)
REPO_ROOT="$GOPATH/src/github.com/diptadas/k8s-extension-apiserver"

pushd ${REPO_ROOT}/hack

rm foo ca.crt ca.key server.crt server.key
kubectl delete -f apiserver.yaml
kubectl delete -f apiservice.yaml
kubectl delete -f apiservice-local.yaml
kubectl delete clusterrolebinding serviceaccounts-cluster-admin

popd
