package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
)

func main() {
	flag.Parse()

	podName := fmt.Sprintf("my-pod-%v", time.Now().Nanosecond())
	fmt.Println("Leader Election for pod:", podName)

	kubeClient, err := getKubeClient()
	if err != nil {
		log.Fatalf(err.Error())
	}

	configMap := &core.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-pod-leader-lock",
			Namespace: metav1.NamespaceDefault,
		},
	}

	resLock := &resourcelock.ConfigMapLock{
		ConfigMapMeta: configMap.ObjectMeta,
		Client:        kubeClient.CoreV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity:      podName,
			EventRecorder: &record.FakeRecorder{},
		},
	}

	stopLeading := make(chan struct{})

	go func() {
		leaderElectionLease := 3 * time.Second

		leaderelection.RunOrDie(leaderelection.LeaderElectionConfig{
			Lock:          resLock,
			LeaseDuration: leaderElectionLease,
			RenewDeadline: leaderElectionLease * 2 / 3,
			RetryPeriod:   leaderElectionLease / 3,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: func(stop <-chan struct{}) {
					go restOfCode(stopLeading)
				},
				OnStoppedLeading: func() {
					stopLeading <- struct{}{}
				},
			},
		})
	}()

	select {}
}

func getKubeClient() (kubernetes.Interface, error) {
	kubeConfigPath := os.Getenv("HOME") + "/.kube/config"
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func restOfCode(stop chan struct{}) {
	fmt.Println("Got leadership, running rest of code...")
	<-stop
	fmt.Println("Lost leadership, exiting rest of code...")
}

/*
go run candidate.go
go run candidate.go
kubectl describe configmap my-pod-leader-lock
kubectl delete configmap my-pod-leader-lock
*/
