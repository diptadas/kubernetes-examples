package util

import (
	"os"
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/clientcmd"
)

func TestParseKubeConfig(t *testing.T) {
	c, err := clientcmd.LoadFromFile(os.Getenv("HOME") + "/.kube/config")
	if err != nil {
		t.Fatal(err)
	}

	s := sets.StringKeySet(c.Contexts)
	t.Log("contexts:", s)
	t.Log("current context:", c.CurrentContext)
	t.Log("default namespace for current context:", c.Contexts[c.CurrentContext].Namespace)
}
