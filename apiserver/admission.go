package apiserver

import (
	"net/http"

	admission "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
)

func Resource() (plural schema.GroupVersionResource, singular string) {
	return schema.GroupVersionResource{
			Group:    "admission.foocontroller.k8s.io",
			Version:  "v1alpha1",
			Resource: "admissionreviews",
		},
		"admissionreview"
}

func Admit(admissionSpec *admission.AdmissionRequest) *admission.AdmissionResponse {
	status := &admission.AdmissionResponse{}
	status.Allowed = false
	status.Result = &metav1.Status{
		Status: metav1.StatusFailure, Code: http.StatusBadRequest, Reason: metav1.StatusReasonBadRequest,
		Message: "force error",
	}
	return status
}

func Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	return nil
}
