package main

import (
	"flag"
	"log"

	"github.com/diptadas/k8s-extension-apiserver/apiserver"
)

func main() {
	flag.Parse()

	stopCh := make(chan struct{})
	defer close(stopCh)

	log.Println("Starting apiserver...")
	if err := apiserver.Run(stopCh); err != nil {
		log.Fatalf("Error running apiserver: %s", err.Error())
	}

	select {}
}
