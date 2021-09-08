#!/usr/bin/env bash
set -xe

# build binary
go build -o foo ../main.go

# build docker image
docker build -t diptadas/foo .

# load docker image to minikube
docker save diptadas/foo | pv | (eval $(minikube docker-env) && docker load)

# create necessary TLS certificates
./onessl create ca-cert
./onessl create server-cert server --domains=foo-operator.default.svc
export SERVICE_SERVING_CERT_CA=$(cat ca.crt | ./onessl base64)
export TLS_SERVING_CERT=$(cat server.crt | ./onessl base64)
export TLS_SERVING_KEY=$(cat server.key | ./onessl base64)
export KUBE_CA=$(./onessl get kube-ca | ./onessl base64)

# create CRD
cat foo-crd.yaml | ./onessl envsubst | kubectl apply -f -

# create operator deployment, service and tls secret
cat operator.yaml | ./onessl envsubst | kubectl apply -f -

# create Validating/Mutating WebhookConfiguration
cat admission.yaml | ./onessl envsubst | kubectl apply -f -

# create APIService
cat apiservice.yaml | ./onessl envsubst | kubectl apply -f -

# cleanup
rm foo ca.crt ca.key server.crt server.key