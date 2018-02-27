package apiserver

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

func Resource() (plural schema.GroupVersionResource, singular string) {
	log.Println("Server Resource")
	return schema.GroupVersionResource{
			Group:    "admission.foocontroller.k8s.io",
			Version:  "v1alpha1",
			Resource: "admissionreviews",
		},
		"admissionreview"
}

func Admit(req *admission.AdmissionRequest) *admission.AdmissionResponse {
	log.Println("Server Admit")
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
	// mutating foo spec
	obj.Spec.ConfigMapName = "k8s-" + obj.Spec.ConfigMapName
	patch := `[{ "op": "replace", "path": "/spec/configMapName", "value": ` + obj.Spec.ConfigMapName + `}]`
	return &admission.AdmissionResponse{
		Allowed: true,
		Patch:   []byte(patch),
	}
}

func Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	log.Println("Server Initialize")
	return nil
}
