package main

import (
	"fmt"
	"log"

	"k8s.io/apimachinery/pkg/labels"
)

func main() {
	str := "aa=a,bb=b"
	m := map[string]string{
		"aa": "a",
		"bb": "b",
		"cc": "c",
	}

	selector, err := labels.Parse(str)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(selector.Matches(labels.Set(m)))
	fmt.Println(selector.String())
}
