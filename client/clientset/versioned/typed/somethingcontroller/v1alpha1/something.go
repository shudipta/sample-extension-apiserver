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

package v1alpha1

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	v1alpha1 "sample-extension-apiserver/apis/somethingcontroller/v1alpha1"
	scheme "sample-extension-apiserver/client/clientset/versioned/scheme"
)

// SomethingsGetter has a method to return a SomethingInterface.
// A group's client should implement this interface.
type SomethingsGetter interface {
	Somethings(namespace string) SomethingInterface
}

// SomethingInterface has methods to work with Something resources.
type SomethingInterface interface {
	Create(*v1alpha1.Something) (*v1alpha1.Something, error)
	Update(*v1alpha1.Something) (*v1alpha1.Something, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.Something, error)
	List(opts v1.ListOptions) (*v1alpha1.SomethingList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Something, err error)
	SomethingExpansion
}

// somethings implements SomethingInterface
type somethings struct {
	client rest.Interface
	ns     string
}

// newSomethings returns a Somethings
func newSomethings(c *SomethingcontrollerV1alpha1Client, namespace string) *somethings {
	return &somethings{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the something, and returns the corresponding something object, and an error if there is any.
func (c *somethings) Get(name string, options v1.GetOptions) (result *v1alpha1.Something, err error) {
	result = &v1alpha1.Something{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("somethings").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Somethings that match those selectors.
func (c *somethings) List(opts v1.ListOptions) (result *v1alpha1.SomethingList, err error) {
	result = &v1alpha1.SomethingList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("somethings").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested somethings.
func (c *somethings) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("somethings").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a something and creates it.  Returns the server's representation of the something, and an error, if there is any.
func (c *somethings) Create(something *v1alpha1.Something) (result *v1alpha1.Something, err error) {
	result = &v1alpha1.Something{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("somethings").
		Body(something).
		Do().
		Into(result)
	return
}

// Update takes the representation of a something and updates it. Returns the server's representation of the something, and an error, if there is any.
func (c *somethings) Update(something *v1alpha1.Something) (result *v1alpha1.Something, err error) {
	result = &v1alpha1.Something{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("somethings").
		Name(something.Name).
		Body(something).
		Do().
		Into(result)
	return
}

// Delete takes name of the something and deletes it. Returns an error if one occurs.
func (c *somethings) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("somethings").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *somethings) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("somethings").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched something.
func (c *somethings) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.Something, err error) {
	result = &v1alpha1.Something{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("somethings").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
