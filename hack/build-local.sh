#!/usr/bin/env bash
set -xe

GOPATH=$(go env GOPATH)
REPO_ROOT="$GOPATH/src/github.com/diptadas/k8s-extension-apiserver"

pushd ${REPO_ROOT}/hack

# create necessary TLS certificates
./onessl create ca-cert
./onessl create server-cert server --domains=foo-apiserver.default.svc
export SERVICE_SERVING_CERT_CA=$(cat ca.crt | ./onessl base64)
export TLS_SERVING_CERT=$(cat server.crt | ./onessl base64)
export TLS_SERVING_KEY=$(cat server.key | ./onessl base64)
export KUBE_CA=$(./onessl get kube-ca | ./onessl base64)

# !!! WARNING !!! Never do this in prod cluster
kubectl create clusterrolebinding serviceaccounts-cluster-admin --clusterrole=cluster-admin --user=system:anonymous || true

# create APIService
kubectl apply -f apiservice-local.yaml

popd

# run locally
go run main.go
