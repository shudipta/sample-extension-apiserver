package something

import (
	"fmt"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"

	"sample-extension-apiserver/apis/somethingcontroller/v1alpha1"
	genericapirequest "k8s.io/apiserver/pkg/endpoints/request"
)

func NewStrategy(typer runtime.ObjectTyper) somethingStrategy {
	fmt.Println("NewStrategy")
	return somethingStrategy{typer, names.SimpleNameGenerator}
}

// GetAttrs returns labels.Set, fields.Set, the presence of Initializers if any
// and error in case the given runtime.Object is not a Flunder
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, bool, error) {
	fmt.Println("GetAttrs")
	apiserver, ok := obj.(*v1alpha1.Something)
	if !ok {
		return nil, nil, false, fmt.Errorf("given object is not a SecDb")
	}
	return labels.Set(apiserver.ObjectMeta.Labels), SelectableFields(apiserver), apiserver.Initializers != nil, nil
}

func MatchSomething(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

// SelectableFields returns a field set that represents the object.
func SelectableFields(obj *v1alpha1.Something) fields.Set {
	return generic.ObjectMetaFieldsSet(&obj.ObjectMeta, true)
}

type somethingStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

func (somethingStrategy) NamespaceScoped() bool {
	return true
}
func (somethingStrategy) PrepareForCreate(ctx genericapirequest.Context, obj runtime.Object) {
	fmt.Println("PrepareForCreate")
}

func (somethingStrategy) PrepareForUpdate(ctx genericapirequest.Context, obj, old runtime.Object) {
	fmt.Println("PrepareForUpdate")
}

func (somethingStrategy) Validate(ctx genericapirequest.Context, obj runtime.Object) field.ErrorList {
	fmt.Println("Validate")
	return field.ErrorList{}
}

func (somethingStrategy) AllowCreateOnUpdate() bool {
	fmt.Println("AllowCreateOnUpdate")
	return false
}

func (somethingStrategy) AllowUnconditionalUpdate() bool {
	fmt.Println("AllowUnconditionalUpdate")
	return false
}

func (somethingStrategy) Canonicalize(obj runtime.Object) {
	fmt.Println("Canonicalize")
}

func (somethingStrategy) ValidateUpdate(ctx genericapirequest.Context, obj, old runtime.Object) field.ErrorList {
	fmt.Println("Canonicalize")
	return field.ErrorList{}
}