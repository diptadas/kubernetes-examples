package admission

import (
	"net/http"

	"encoding/json"
	api "k8s-extension-apiserver/apis/foocontroller/v1alpha1"
	"log"

	admission "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

type FooValidator struct{}

func (*FooValidator) ValidatingResource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    "validation.foocontroller.k8s.io",
			Version:  "v1alpha1",
			Resource: "admissionreviews",
		},
		"admissionreview"
}

func (*FooValidator) Validate(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	log.Println("FooValidator: " + req.Operation)

	obj := &api.Foo{}
	if err := json.Unmarshal(req.Object.Raw, obj); err != nil {
		return &admission.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Status: metav1.StatusFailure, Code: http.StatusBadRequest, Reason: metav1.StatusReasonBadRequest,
				Message: "invalid foo object",
			},
		}
	}

	if obj.Annotations == nil || obj.Annotations["initial-configmap"] != obj.Spec.ConfigMapName {
		return &admission.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Status: metav1.StatusFailure, Code: http.StatusBadRequest, Reason: metav1.StatusReasonBadRequest,
				Message: "invalid configMapName",
			},
		}
	}

	return &admission.AdmissionResponse{Allowed: true}
}

func (*FooValidator) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	log.Println("FooValidator: Initialize")
	return nil
}
