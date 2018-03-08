#!/usr/bin/env bash

#set -x
#
GOPATH=$(go env GOPATH)
PACKAGE_NAME=sample-extension-apiserver
REPO_ROOT="$GOPATH/src/$PACKAGE_NAME"
#DOCKER_REPO_ROOT="/go/src/$PACKAGE_NAME"
#DOCKER_CODEGEN_PKG="/go/src/k8s.io/code-generator"
#
pushd $REPO_ROOT

rm -rf "$REPO_ROOT"/apis/somethingcontroller/v1alpha1/*.generated.go
rm -rf "$REPO_ROOT"/apis/somethingcontroller/install/*.generated.go

mkdir apis
mkdir apis/somethingcontroller
mkdir apis/somethingcontroller/v1alpha1
mkdir apis/somethingcontroller/install

chmod +x hack/gen_ctl_register.sh
hack/gen_ctl_register.sh

chmod +x hack/gen_ver_doc.sh
hack/gen_ver_doc.sh

chmod +x hack/gen_ver_register.sh
hack/gen_ver_register.sh

chmod +x hack/gen_ver_types.sh
hack/gen_ver_types.sh

#chmod +x hack/gen_ver_zz.sh
#hack/gen_ver_zz.sh

chmod +x hack/gen_install.sh
hack/gen_install.sh

chmod +x hack/update-codegen.sh
hack/update-codegen.sh
#
#rm -rf "$REPO_ROOT"/apis/something-controller/v1alpha1/*.generated.go
##
##docker run --rm -ti -u $(id -u):$(id -g) \
##  -v "$REPO_ROOT":"$DOCKER_REPO_ROOT" \
##  -w "$DOCKER_REPO_ROOT" \
##  appscode/gengo:release-1.9 "$DOCKER_CODEGEN_PKG"/generate-internal-groups.sh deepcopy \
##  "$PACKAGE_NAME/client" \
##  "$PACKAGE_NAME/apis" \
##  "$PACKAGE_NAME/apis" \
##  something-controller:v1alpha1 \
##  --go-header-file "$DOCKER_REPO_ROOT/hack/gengo/boilerplate.go.txt"
#
##
#mkdir pkg
#mkdir pkg/apiserver
#touch pkg/apiserver/apiserver.go
#echo "package apiserver" > pkg/apiserver/apiserver.go
#
#mkdir cmd/server
##chmod +x hack/gen_cmd_server.sh
##hack/gen_cmd_server.sh
#


popd
