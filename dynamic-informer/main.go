package main

import (
	"fmt"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	metacontrollerapi "k8s.io/metacontroller/apis/metacontroller/v1alpha1"
	dynamicclientset "k8s.io/metacontroller/dynamic/clientset"
	dynamicdiscovery "k8s.io/metacontroller/dynamic/discovery"
	dynamicinformer "k8s.io/metacontroller/dynamic/informer"
)

// list of resources to watch
var myResources = []metacontrollerapi.ResourceRule{
	{
		APIVersion: "v1",
		Resource:   "pods",
	},
	{
		APIVersion: "v1",
		Resource:   "services",
	},
	{
		APIVersion: "mycrd.k8s.io/v1alpha1",
		Resource:   "foos",
	},
}

func main() {
	// create rest-config from kube-config file
	kubeConfigPath := os.Getenv("HOME") + "/.kube/config"
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		panic(err)
	}

	// resync periods
	discoveryInterval := 5 * time.Second
	informerRelist := 5 * time.Minute

	// Periodically refresh discovery to pick up newly-installed resources.
	dc := discovery.NewDiscoveryClientForConfigOrDie(config)
	resources := dynamicdiscovery.NewResourceMap(dc)
	// We don't care about stopping this cleanly since it has no external effects.
	resources.Start(discoveryInterval)

	// Create dynamic clientset (factory for dynamic clients).
	dynClient := dynamicclientset.New(config, resources)
	// Create dynamic informer factory (for sharing dynamic informers).
	dynInformers := dynamicinformer.NewSharedInformerFactory(dynClient, informerRelist)

	fmt.Println("waiting for sync")
	if !resources.HasSynced() {
		time.Sleep(time.Second)
	}

	// create informer for all resources
	var informers []*dynamicinformer.ResourceInformer
	for _, res := range myResources {
		informer, err := dynInformers.Resource(res.APIVersion, res.Resource)
		if err != nil {
			panic("can't create informer for resource: " + err.Error())
		}
		fmt.Println("created informar for resource:", res.APIVersion, res.Resource)
		informers = append(informers, informer)
	}

	// common handler for all events
	printObjectMeta := func(event string, obj interface{}) {
		o := obj.(*unstructured.Unstructured)
		fmt.Printf("%s %s:%s:%s:%s\n", event, o.GetAPIVersion(), o.GetKind(), o.GetNamespace(), o.GetName())
	}
	handler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			printObjectMeta("added", obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			printObjectMeta("updated", newObj)
		},
		DeleteFunc: func(obj interface{}) {
			printObjectMeta("deleted", obj)
		},
	}

	// add event handler for informers
	for _, informer := range informers {
		informer.Informer().AddEventHandler(handler)
	}

	select {}
}
