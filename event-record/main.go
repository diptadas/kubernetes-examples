package main

import (
	"context"
	"fmt"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"time"

	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/tools/reference"
)

func main() {
	kubeConfigPath := os.Getenv("HOME") + "/.kube/config"
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		panic(err)
	}
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	log.Println("Creating new configmap as reference object")
	referenceObject, err := kubeClient.CoreV1().ConfigMaps(metav1.NamespaceDefault).Create(
		context.Background(),
		&core.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "example-configmap",
				Namespace: metav1.NamespaceDefault,
			},
		},
		metav1.CreateOptions{},
	)
	if err != nil && kerr.IsAlreadyExists(err) {
		panic(err)
	}

	e := eventer{
		client:    kubeClient,
		component: "golang-examples",
	}

	log.Println("Recording events using broadcaster")
	e.newEventRecorder().Event(referenceObject, core.EventTypeNormal, "event-test-1", "new event is recorded")
	time.Sleep(time.Second) // time to complete event

	log.Println("Creating events directly")
	event, err := e.createEvent(referenceObject, core.EventTypeNormal, "event-test-2", "new event is recorded")
	if err != nil {
		panic(err)
	}
	log.Println("Event recorded:", event.Name)
}

type eventer struct {
	client    kubernetes.Interface
	component string
}

func (e eventer) newEventRecorder() record.EventRecorder {
	// Event Broadcaster
	broadcaster := record.NewBroadcaster()
	broadcaster.StartEventWatcher(
		func(event *core.Event) {
			if _, err := e.client.CoreV1().Events(event.Namespace).Create(context.Background(), event, metav1.CreateOptions{}); err != nil {
				log.Println(err)
			} else {
				log.Println("Event recorded:", event.Name)
			}
		},
	)
	// Event Recorder
	return broadcaster.NewRecorder(scheme.Scheme, core.EventSource{Component: e.component})
}

func (e eventer) createEvent(object runtime.Object, eventType, reason, message string) (*core.Event, error) {
	ref, err := reference.GetReference(scheme.Scheme, object)
	if err != nil {
		return nil, err
	}

	t := metav1.Time{Time: time.Now()}

	return e.client.CoreV1().Events(ref.Namespace).Create(
		context.Background(),
		&core.Event{
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
			Source:         core.EventSource{Component: e.component},
		},
		metav1.CreateOptions{},
	)
}

// go test -v -count=1 -run TestEventRecord ./event_record
// kubectl describe configmap example-configmap
// kubectl get events
