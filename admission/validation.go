package admission

import (
	"net/http"

	"encoding/json"
	api "k8s-admission-webhook/apis/foocontroller/v1alpha1"
	"log"

	admission "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

type FooValidator struct{}

func (*FooValidator) ValidatingResource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    "admission.foocontroller.k8s.io",
			Version:  "v1alpha1",
			Resource: "validations",
		},
		"validation"
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

	if req.Operation == admission.Delete {
		if obj.Annotations != nil && obj.Annotations["deny-delete"] == "true" {
			return &admission.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Status: metav1.StatusFailure, Code: http.StatusBadRequest, Reason: metav1.StatusReasonBadRequest,
					Message: "force denied delete",
				},
			}
		}
		return &admission.AdmissionResponse{Allowed: true}
	}

	oldObj := &api.Foo{}
	if err := json.Unmarshal(req.OldObject.Raw, oldObj); err != nil {
		return &admission.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Status: metav1.StatusFailure, Code: http.StatusBadRequest, Reason: metav1.StatusReasonBadRequest,
				Message: "invalid old foo object",
			},
		}
	}

	// deny update if configMapName changed
	if obj.Spec.ConfigMapName != oldObj.Spec.ConfigMapName {
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
