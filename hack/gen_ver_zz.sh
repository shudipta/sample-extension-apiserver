#!/usr/bin/env bash

pushd $GOPATH/src/sample-extension-apiserver/apis/somethingcontroller/v1alpha1

touch zz_generated.deepcopy.go
echo "// +build !ignore_autogenerated

/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the \"License\");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an \"AS IS\" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This file was autogenerated by deepcopy-gen. Do not edit it manually!

package v1alpha1

import (
	runtime \"k8s.io/apimachinery/pkg/runtime\"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Something) DeepCopyInto(out *Something) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Something.
func (in *Something) DeepCopy() *Something {
	if in == nil {
		return nil
	}
	out := new(Something)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Something) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SomethingList) DeepCopyInto(out *SomethingList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Something, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SomethingList.
func (in *SomethingList) DeepCopy() *SomethingList {
	if in == nil {
		return nil
	}
	out := new(SomethingList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SomethingList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SomethingSpec) DeepCopyInto(out *SomethingSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		if *in == nil {
			*out = nil
		} else {
			*out = new(int32)
			**out = **in
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SomethingSpec.
func (in *SomethingSpec) DeepCopy() *SomethingSpec {
	if in == nil {
		return nil
	}
	out := new(SomethingSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SomethingStatus) DeepCopyInto(out *SomethingStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SomethingStatus.
func (in *SomethingStatus) DeepCopy() *SomethingStatus {
	if in == nil {
		return nil
	}
	out := new(SomethingStatus)
	in.DeepCopyInto(out)
	return out
}
" > zz_generated.deepcopy.go

popd
