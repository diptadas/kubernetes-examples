package admission

import (
	"net/http"

	"encoding/json"
	"log"

	api "github.com/diptadas/k8s-admission-webhook/apis/foocontroller/v1alpha1"

	admission "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

type FooMutator struct{}

func (*FooMutator) MutatingResource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    "admission.foocontroller.k8s.io",
			Version:  "v1alpha1",
			Resource: "mutations",
		},
		"mutation"
}

func (*FooMutator) Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	log.Println("FooMutator: " + req.Operation)

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

	if obj.Spec.ConfigMapName == "" {
		return &admission.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Status: metav1.StatusFailure, Code: http.StatusBadRequest, Reason: metav1.StatusReasonBadRequest,
				Message: "configMapName not specified",
			},
		}
	}

	// mutating foo: add "deny-delete" annotation
	patch := `[{"op": "add", "path": "/metadata/annotations/deny-delete", "value": "true"}]`
	return &admission.AdmissionResponse{
		Allowed: true,
		Patch:   []byte(patch),
	}
}

func (*FooMutator) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	log.Println("FooMutator: Initialize")
	return nil
}
