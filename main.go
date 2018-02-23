package main

import (
	"flag"
	clientset "k8s-extension-apiserver/client/clientset/versioned"
	"k8s-extension-apiserver/controller"
	"log"

	"k8s-extension-apiserver/apiserver"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	masterURL  string
	kubeconfig string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "/home/dipta/.kube/config", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}

func main() {
	flag.Parse()

	stopCh := make(chan struct{})
	defer close(stopCh)

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	fooClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building foo clientset: %s", err.Error())
	}

	controller := controller.NewController(kubeClient, fooClient)
	go func() {
		log.Println("Starting controller...")
		if err = controller.Run(stopCh); err != nil {
			log.Fatalf("Error running controller: %s", err.Error())
		}
	}()

	go func() {
		log.Println("Starting apiserver...")
		if err = apiserver.Run(cfg, stopCh); err != nil {
			log.Fatalf("Error running apiserver: %s", err.Error())
		}
	}()

	select {}
}
