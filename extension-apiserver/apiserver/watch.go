package apiserver

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/diptadas/kubernetes-examples/extension-apiserver/apis/foocontroller/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
)

type FooWatcher struct {
	result  chan watch.Event
	Stopped bool
	sync.Mutex
}

func (f *FooWatcher) Stop() {
	f.Lock()
	defer f.Unlock()
	if !f.Stopped {
		log.Println("Stopping foo watcher...")
		close(f.result)
		f.Stopped = true
	}
}

func (f *FooWatcher) ResultChan() <-chan watch.Event {
	return f.result
}

func (f *FooWatcher) Add(obj runtime.Object) {
	log.Println("Adding foo...")
	f.result <- watch.Event{watch.Added, obj}
}

func NewFooWatcher() *FooWatcher {
	fw := &FooWatcher{
		result: make(chan watch.Event),
	}
	log.Println("Starting foo watcher...")
	go fw.RunWatcher()
	return fw
}

func (f *FooWatcher) RunWatcher() {
	for i := 0; !f.Stopped; i++ {
		f.Add(&v1alpha1.Foo{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "foocontroller.k8s.io/v1alpha1",
				Kind:       "Foo",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("myfoo-%d", i),
				Namespace: "default",
			},
			Spec: "do-not-care",
		})
		time.Sleep(time.Second * 5)
	}
}
