#!/usr/bin/env bash
set -xe

GOPATH=$(go env GOPATH)
REPO_ROOT="$GOPATH/src/github.com/diptadas/k8s-extension-apiserver"

function cleanup {
  ${REPO_ROOT}/hack/cleanup.sh
}
trap cleanup EXIT

pushd ${REPO_ROOT}/hack

# build binary
go build -o foo ../main.go

# build docker image
docker build -t diptadas/foo .

# load docker image to minikube
docker save diptadas/foo | pv | (eval $(minikube docker-env) && docker load)

# create necessary TLS certificates
./onessl create ca-cert
./onessl create server-cert server --domains=foo-apiserver.default.svc
export SERVICE_SERVING_CERT_CA=$(cat ca.crt | ./onessl base64)
export TLS_SERVING_CERT=$(cat server.crt | ./onessl base64)
export TLS_SERVING_KEY=$(cat server.key | ./onessl base64)
export KUBE_CA=$(./onessl get kube-ca | ./onessl base64)

# !!! WARNING !!! Never do this in prod cluster
kubectl create clusterrolebinding serviceaccounts-cluster-admin --clusterrole=cluster-admin --group=system:serviceaccounts || true

# create apiserver deployment and tls secret
cat apiserver.yaml | ./onessl envsubst | kubectl apply -f -

# create APIService
cat apiservice.yaml | ./onessl envsubst | kubectl apply -f -

popd
