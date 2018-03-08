package apiserver

import (
	"errors"
	"log"

	"sample-extension-apiserver/apis/somethingcontroller/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	apirequest "k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/rest"
)

type REST struct{}

var _ rest.Getter = &REST{}
var _ rest.GroupVersionKindProvider = &REST{}

func NewREST() *REST {
	return &REST{}
}

func (r *REST) New() runtime.Object {
	return &v1alpha1.Something{}
}

func (r *REST) GroupVersionKind(containingGV schema.GroupVersion) schema.GroupVersionKind {
	return v1alpha1.SchemeGroupVersion.WithKind("Foo")
}

func intPtr(i int32) *int32 {
	return &i
}

func (r *REST) Get(ctx apirequest.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	log.Println("Get...")

	ns, ok := apirequest.NamespaceFrom(ctx)
	if !ok {
		return nil, errors.New("missing namespace")
	}
	if len(name) == 0 {
		return nil, errors.New("missing search query")
	}

	resp := &v1alpha1.Something{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "somethingcontroller.kube-ac.com/v1alpha1",
			Kind:       "Something",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Spec: v1alpha1.SomethingSpec{
			DeploymentName: "something-exmp",
			Replicas: intPtr(1),
		},
	}

	return resp, nil
}