package util

import (
	"testing"

	"k8s.io/apimachinery/pkg/labels"
)

func TestParseSelector(t *testing.T) {
	selectorStr := "aa=a,bb=b"
	labelsMap := map[string]string{
		"aa": "a",
		"bb": "b",
		"cc": "c",
	}

	selector, err := labels.Parse(selectorStr)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("is matched:", selector.Matches(labels.Set(labelsMap)))
	t.Log("selector:", selector.String())
}
