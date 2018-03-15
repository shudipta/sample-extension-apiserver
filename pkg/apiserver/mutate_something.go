package apiserver

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"k8s.io/client-go/rest"
	"log"
)

type SomethingMutaionHook struct {}

func (*SomethingMutaionHook) MutatingResource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
		Group:    "admission.somethingcontroller.kube-ac.com",
		Version:  "v1alpha1",
		Resource: "mutations",
	},
		"mutation"
}

func (*SomethingMutaionHook) Admit(
	req *admissionv1beta1.AdmissionRequest) *admissionv1beta1.AdmissionResponse {
	log.Println("---------mutating...........")

	mutatingObjectMeta := &NamedThing{}
	err := json.Unmarshal(req.Object.Raw, mutatingObjectMeta)
	if err != nil {
		log.Println("---------invalid obj...........")

		return &admissionv1beta1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Status: metav1.StatusFailure, Code: http.StatusBadRequest, Reason: metav1.StatusReasonBadRequest,
				Message: err.Error(),
			},
		}
	}

	if req.Operation == admissionv1beta1.Create {
		if _, ok := mutatingObjectMeta.Annotations["sample-label"]; !ok {
			log.Println("---------mutating the something obj...........")

			patch := `[{"op": "add", "path": "/metadata/annotations/sample-label", "value": "true"}]`
			return &admissionv1beta1.AdmissionResponse{
				Allowed: true,
				Patch:   []byte(patch),
			}
		}
	}

	return &admissionv1beta1.AdmissionResponse{Allowed: true}
}

func (*SomethingMutaionHook) Initialize(config *rest.Config, stopCh <-chan struct{}) error {
	log.Println("SomethingValidator: Initialize")
	return nil
}
