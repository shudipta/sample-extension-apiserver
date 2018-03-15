#!/usr/bin/env bash
#set -xe
GOPATH=$(go env GOPATH)
PACKAGE_NAME=sample-extension-apiserver
REPO_ROOT="$GOPATH/src/$PACKAGE_NAME"
#DOCKER_REPO_ROOT="/go/src/$PACKAGE_NAME"
#DOCKER_CODEGEN_PKG="/go/src/k8s.io/code-generator"
#
pushd $REPO_ROOT

rm -rf ./main ./hack/server-image/sample-extension-apiserver

kubectl delete -f hack/deploy/crd.yaml

kubectl delete -f hack/deploy/sa.yaml

kubectl delete -f hack/deploy/clrb.yaml -n kube-system
kubectl delete -f hack/deploy/rb.yaml -n kube-system

kubectl delete -f hack/deploy/rc.yaml
kubectl delete -f hack/deploy/svc.yaml
kubectl delete -f hack/deploy/config.yaml

kubectl delete -f hack/deploy/admission.yaml
kubectl delete -f hack/deploy/apiservice.yaml

kubectl delete -f hack/deploy/something.yaml

kubectl delete ns kube-ac

popd