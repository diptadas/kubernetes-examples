package main

import (
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
)

func main() {
	c, err := clientcmd.LoadFromFile(os.Getenv("HOME") + "/.kube/config")
	if err != nil {
		panic(err)
	}

	s := sets.StringKeySet(c.Contexts)
	log.Println("contexts:", s)
	log.Println("current context:", c.CurrentContext)
	log.Println("default namespace for current context:", c.Contexts[c.CurrentContext].Namespace)
}
