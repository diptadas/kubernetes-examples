package main

import (
	"k8s-admission-webhook/apiserver"
	"log"
)

func main() {
	stopCh := make(chan struct{})
	defer close(stopCh)

	log.Println("Starting apiserver...")
	if err := apiserver.Run(stopCh); err != nil {
		log.Fatalf("Error running apiserver: %s", err.Error())
	}

	select {}
}
