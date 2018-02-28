package main

import (
	"fmt"
	"log"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	c, err := clientcmd.LoadFromFile("/home/dipta/.kube/config")
	if err != nil {
		log.Panic(err)
	}

	s := sets.StringKeySet(c.Contexts)
	fmt.Println(s)

	fmt.Println(c.CurrentContext)
	fmt.Println(c.Contexts[c.CurrentContext].Namespace)
}
