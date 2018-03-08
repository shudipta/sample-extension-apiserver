/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fake

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
	v1alpha1 "sample-extension-apiserver/apis/somethingcontroller/v1alpha1"
)

// FakeSomethings implements SomethingInterface
type FakeSomethings struct {
	Fake *FakeSomethingcontrollerV1alpha1
	ns   string
}

var somethingsResource = schema.GroupVersionResource{Group: "somethingcontroller.kube-ac.com", Version: "v1alpha1", Resource: "somethings"}

var somethingsKind = schema.GroupVersionKind{Group: "somethingcontroller.kube-ac.com", Version: "v1alpha1", Kind: "Something"}

// Get takes name of the something, and returns the corresponding something object, and an error if there is any.
func (c *FakeSomethings) Get(name string, options v1.GetOptions) (result *v1alpha1.Something, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(somethingsResource, c.ns, name), &v1alpha1.Something{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Something), err
}

// List takes label and field selectors, and returns the list of Somethings that match those selectors.
func (c *FakeSomethings) List(opts v1.ListOptions) (result *v1alpha1.SomethingList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(somethingsResource, somethingsKind, c.ns, opts), &v1alpha1.SomethingList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.SomethingList{}
	for _, item := range obj.(*v1alpha1.SomethingList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested somethings.
func (c *FakeSomethings) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(somethingsResource, c.ns, opts))

}

// Create takes the representation of a something and creates it.  Returns the server's representation of the something, and an error, if there is any.
func (c *FakeSomethings) Create(something *v1alpha1.Something) (result *v1alpha1.Something, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(somethingsResource, c.ns, something), &v1alpha1.Something{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Something), err
}

// Update takes the representation of a something and updates it. Returns the server's representation of the something, and an error, if there is any.
func (c *FakeSomethings) Update(something *v1alpha1.Something) (result *v1alpha1.Something, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(somethingsResource, c.ns, something), &v1alpha1.Something{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Something), err
}

// Delete takes name of the something and deletes it. Returns an error if one occurs.
func (c *FakeSomethings) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(somethingsResource, c.ns, name), &v1alpha1.Something{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSomethings) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(somethingsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.SomethingList{})
	return err
}

// Patch applies the patch and returns the patched something.
func (c *FakeSomethings) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Something, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(somethingsResource, c.ns, name, data, subresources...), &v1alpha1.Something{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Something), err
}
