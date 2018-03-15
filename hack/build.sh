#!/usr/bin/env bash
#set -xe
GOPATH=$(go env GOPATH)
PACKAGE_NAME=sample-extension-apiserver
REPO_ROOT="$GOPATH/src/$PACKAGE_NAME"
#DOCKER_REPO_ROOT="/go/src/$PACKAGE_NAME"
#DOCKER_CODEGEN_PKG="/go/src/k8s.io/code-generator"
#
pushd $REPO_ROOT

rm ./ca.crt ./ca.key ./server.crt ./server.key
rm ./main
# build binary
export GOOS=linux; go build *.go
cp ./main ./hack/server-image/sample-extension-apiserver
docker build -t shudipta/sample-extension-apiserver:v2 ./hack/server-image
docker push shudipta/sample-extension-apiserver:v2

#hack/deploy/onessl create ca-cert
#hack/deploy/onessl create server-cert server --domains=svc-apiserver.kube-ac.svc
#export SERVICE_SERVING_CERT_CA=$(cat ca.crt | hack/deploy/onessl base64)
#export TLS_SERVING_CERT=$(cat server.crt | hack/deploy/onessl base64)
#export TLS_SERVING_KEY=$(cat server.key | hack/deploy/onessl base64)
#export KUBE_CA=$(hack/deploy/onessl get kube-ca | hack/deploy/onessl base64)
#echo $KUBE_CA
##export SERVICE_SERVING_CERT_CA=$(cat ca.crt | ./hack/deploy/onessl base64)
#export TLS_SERVING_CERT=$(cat $REPO_ROOT/apiserver.local.config/certificates/apiserver.crt | hack/deploy/onessl base64)
#export TLS_SERVING_KEY=$(cat $REPO_ROOT/apiserver.local.config/certificates/apiserver.key | hack/deploy/onessl base64)

kubectl create ns kube-ac
kubectl apply -f hack/deploy/crd.yaml
#
kubectl apply -f hack/deploy/sa.yaml -n kube-ac

kubectl apply -f hack/deploy/clrb.yaml -n kube-system
kubectl apply -f hack/deploy/rb.yaml -n kube-system

kubectl apply -f hack/deploy/rc.yaml -n kube-ac
kubectl apply -f hack/deploy/svc.yaml -n kube-ac

kubectl apply -f hack/deploy/admission.yaml --validate=false
kubectl apply -f hack/deploy/apiservice.yaml

#kubectl apply -f hack/deploy/something.yaml
#kubectl create rolebinding -n kube-system \
#    extension-apiserver-authentication-reader \
#    --role=extension-apiserver-authentication-reader \
#    --serviceaccount=kube-ac:sa-apiserver

#cat hack/deploy/config.yaml | hack/deploy/onessl envsubst | kubectl apply -f -
#cat hack/deploy/admission.yaml | hack/deploy/onessl envsubst | kubectl apply -f -
#cat hack/deploy/apiservice.yaml | hack/deploy/onessl envsubst | kubectl apply -f -

popd