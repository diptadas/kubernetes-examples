package main

import (
	"k8s.io/apimachinery/pkg/labels"
	"log"
)

func main() {
	selectorStr := "aa=a,bb=b"
	labelsMap := map[string]string{
		"aa": "a",
		"bb": "b",
		"cc": "c",
	}

	selector, err := labels.Parse(selectorStr)
	if err != nil {
		panic(err)
	}

	log.Println("is matched:", selector.Matches(labels.Set(labelsMap)))
	log.Println("to string:", selector.String())
}
