package leader_election

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/diptadas/kubernetes-examples/util"

	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
)

const (
	leaderElectionConfigMap = "leader-election-lock"
)

func runLeaderElection(podName string, kubeClient kubernetes.Interface) {
	fmt.Println("Leader Election for pod:", podName)

	resLock := &resourcelock.ConfigMapLock{
		Client: kubeClient.CoreV1(),
		ConfigMapMeta: metav1.ObjectMeta{
			Name:      leaderElectionConfigMap,
			Namespace: metav1.NamespaceDefault,
		},
		LockConfig: resourcelock.ResourceLockConfig{
			Identity:      podName,
			EventRecorder: &record.FakeRecorder{},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:          resLock,
		LeaseDuration: 15 * time.Second,
		RenewDeadline: 10 * time.Second,
		RetryPeriod:   2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				fmt.Println("Got leadership for pod:", podName)
				time.Sleep(5 * time.Second)
				cancel() // release leadership after 5 sec
			},
			OnStoppedLeading: func() {
				fmt.Println("Lost leadership for pod:", podName)
			},
		},
	})

	fmt.Println("Closing Leader Election for pod:", podName)
}

func TestLeaderElection(t *testing.T) {
	kubeClient, err := util.GetKubeClient()
	if err != nil {
		t.Fatal(err)
	}

	// run two candidate in separate go routine
	go runLeaderElection("pod-1", kubeClient)
	time.Sleep(1 * time.Second)
	runLeaderElection("pod-2", kubeClient)

	// cleanup configMap
	err = kubeClient.CoreV1().ConfigMaps(metav1.NamespaceDefault).Delete(leaderElectionConfigMap, &metav1.DeleteOptions{})
	if err != nil && kerr.IsNotFound(err) {
		t.Fatal(err)
	}
}

// go test -v -count=1 -run TestLeaderElection ./leader_election
