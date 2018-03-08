#!/usr/bin/env bash

pushd $GOPATH/src/sample-extension-apiserver/apis/somethingcontroller/v1alpha1

touch doc.go
echo "// +k8s:deepcopy-gen=package,register

// Package v1alpha1 is the alpha release of v1 version of the API.
// +groupName=somethingcontroller.kube-ac.com
package v1alpha1
" > doc.go

popd
