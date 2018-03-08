#!/usr/bin/env bash

pushd $GOPATH/src/sample-extension-apiserver/apis/somethingcontroller

touch register.go
echo "package somethingcontroller

const GroupName string = \"somethingcontroller.kube-ac.com\"
" > register.go

popd
