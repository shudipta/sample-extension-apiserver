#!/usr/bin/env bash
#set -xe
GOPATH=$(go env GOPATH)
PACKAGE_NAME=sample-extension-apiserver
REPO_ROOT="$GOPATH/src/$PACKAGE_NAME"

pushd $REPO_ROOT

# build binary
export GOOS=linux; go build .
cp ./sample-extension-apiserver ./hack/server-image/sample-extension-apiserver
docker build -t shudipta/sample-extension-apiserver:latest ./hack/server-image
docker push shudipta/sample-extension-apiserver:latest

kubectl apply ns kube-ac

kubectl apply -f hack/deploy/sa.yaml -n kube-ac

kubectl apply -f hack/deploy/clrb.yaml -n kube-system
kubectl apply -f hack/deploy/rb.yaml -n kube-system

kubectl apply -f hack/deploy/rc.yaml -n kube-ac
kubectl apply -f hack/deploy/svc.yaml -n kube-ac

kubectl apply -f hack/deploy/apiservice.yaml

kubectl apply -f hack/deploy/something.yaml

popd