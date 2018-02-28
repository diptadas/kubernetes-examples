package main

import (
	"fmt"
	"log"
	"os"
	"time"

	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/tools/reference"
)

func main() {
	kubeClient, err := getKubeClient()
	if err != nil {
		log.Fatal(err.Error())
	}

	resource := &core.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-configmap",
			Namespace: metav1.NamespaceDefault,
		},
	}

	log.Println("Deleting old resource")
	err = kubeClient.CoreV1().ConfigMaps(resource.Namespace).Delete(resource.Name, &metav1.DeleteOptions{})
	if err != nil && !kerr.IsNotFound(err) {
		log.Fatal(err.Error())
	}

	log.Println("Creating new resource")
	resource, err = kubeClient.CoreV1().ConfigMaps(resource.Namespace).Create(resource)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("Recording events using broadcaster")
	recorder := newEventRecorder(kubeClient, "golang-examples")
	recorder.Event(resource, core.EventTypeNormal, "event-test-1", "new event is recorded")
	time.Sleep(time.Second) // time to complete event

	log.Println("Creating events directly")
	event, err := createEvent(kubeClient, "golang-examples", resource, core.EventTypeNormal, "event-test-2", "new event is recorded")
	if err != nil {
		log.Fatal(err.Error())
	} else {
		log.Println("Event recorded:", event.Name)
	}
}

func getKubeClient() (kubernetes.Interface, error) {
	kubeConfigPath := os.Getenv("HOME") + "/.kube/config"
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func newEventRecorder(client kubernetes.Interface, component string) record.EventRecorder {
	// Event Broadcaster
	broadcaster := record.NewBroadcaster()
	broadcaster.StartEventWatcher(
		func(event *core.Event) {
			if _, err := client.CoreV1().Events(event.Namespace).Create(event); err != nil {
				log.Println(err)
			} else {
				log.Println("Event recorded:", event.Name)
			}
		},
	)
	// Event Recorder
	return broadcaster.NewRecorder(scheme.Scheme, core.EventSource{Component: component})
}

func createEvent(client kubernetes.Interface, component string, obj runtime.Object, eventType, reason, message string) (*core.Event, error) {
	ref, err := reference.GetReference(scheme.Scheme, obj)
	if err != nil {
		return nil, err
	}

	t := metav1.Time{Time: time.Now()}

	return client.CoreV1().Events(ref.Namespace).Create(&core.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%v.%x", ref.Name, t.UnixNano()),
			Namespace: ref.Namespace,
		},
		InvolvedObject: *ref,
		Reason:         reason,
		Message:        message,
		FirstTimestamp: t,
		LastTimestamp:  t,
		Count:          1,
		Type:           eventType,
		Source:         core.EventSource{Component: component},
	})
}
