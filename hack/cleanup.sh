#!/usr/bin/env bash
set -x

rm foo ca.crt ca.key server.crt server.key
kubectl delete -f foo-one.yaml
kubectl delete -f foo-crd.yaml
kubectl delete -f operator.yaml
kubectl delete -f admission.yaml
kubectl delete -f apiservice.yaml
