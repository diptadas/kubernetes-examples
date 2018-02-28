package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"k8s.io/api/admissionregistration/v1alpha1"
	apps "k8s.io/api/apps/v1beta1"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ktypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const initializerName = "com.my.initializer"

func main() {
	kubeClient, err := getKubeClient()
	if err != nil {
		log.Fatal(err.Error())
	}

	initConfig := initializerForResources([]string{"configmaps"})
	fmt.Println("Creating InitializerConfigurations:", initConfig.Name)
	if _, err = kubeClient.AdmissionregistrationV1alpha1().InitializerConfigurations().Create(&initConfig); err != nil {
		log.Fatalf(err.Error())
	}
	time.Sleep(3 * time.Second) // time for initializer to be ready

	runController(kubeClient)

	// creating resource
	configMap := core.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-configmap",
			Namespace: metav1.NamespaceDefault,
		},
	}
	fmt.Println("Creating configmap:", configMap.Name)
	if _, err = kubeClient.CoreV1().ConfigMaps(configMap.Namespace).Create(&configMap); err != nil {
		log.Fatalf(err.Error())
	}

	fmt.Println("Getting configmap:", configMap.Name)
	obj, err := kubeClient.CoreV1().ConfigMaps(configMap.Namespace).Get(configMap.Name, metav1.GetOptions{})
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Println("Configmap data:", obj.Data)

	// cleanup
	if err = kubeClient.CoreV1().ConfigMaps(configMap.Namespace).Delete(configMap.Name, &metav1.DeleteOptions{}); err != nil {
		log.Fatalf(err.Error())
	}
	if err = kubeClient.AdmissionregistrationV1alpha1().InitializerConfigurations().Delete(initConfig.Name, &metav1.DeleteOptions{}); err != nil {
		log.Fatalf(err.Error())
	}
}

func getKubeClient() (kubernetes.Interface, error) {
	kubeConfigPath := os.Getenv("HOME") + "/.kube/config"
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func runController(kubeClient kubernetes.Interface) {
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.IncludeUninitialized = true
			return kubeClient.CoreV1().ConfigMaps(core.NamespaceAll).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.IncludeUninitialized = true
			return kubeClient.CoreV1().ConfigMaps(core.NamespaceAll).Watch(options)
		},
	}

	resyncPeriod := 30 * time.Second

	_, controller := cache.NewInformer(
		lw,
		&core.ConfigMap{},
		resyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				configMap := obj.(*core.ConfigMap)
				fmt.Println("Added configmap:", configMap.Name)

				if configMap.GetInitializers() == nil {
					fmt.Println("Already initialized configmap:", configMap.Name)
				} else if configMap.GetInitializers().Pending[0].Name != initializerName {
					fmt.Println("Not my turn to initialize configmap:", configMap.Name)
				} else {
					fmt.Println("Initializing configmap:", configMap.Name)
					if _, err := initializeConfigmap(kubeClient, configMap); err != nil {
						fmt.Printf("Failed to initialize configmap %s, reason %s\n", configMap.Name, err)
					} else {
						fmt.Println("Initialized configmap:", configMap.Name)
					}
				}
			},
		},
	)

	fmt.Println("Running controller...")
	go controller.Run(make(chan struct{}))
}

func initializeConfigmap(kubeClient kubernetes.Interface, configMap *core.ConfigMap) (*core.ConfigMap, error) {
	return patchConfigMap(kubeClient, configMap, func(configMap *core.ConfigMap) *core.ConfigMap {
		pending := configMap.Initializers.Pending
		fmt.Println("Removing pending initializer:", pending[0].Name)
		if len(pending) == 1 {
			configMap.ObjectMeta.Initializers = nil
		} else {
			configMap.ObjectMeta.Initializers.Pending = append(pending[:0], pending[1:]...)
		}

		// adding some data to check letter
		configMap.Data = map[string]string{
			"initialized-by": initializerName,
		}
		return configMap
	})
}

func patchConfigMap(c kubernetes.Interface, cur *core.ConfigMap, transform func(*core.ConfigMap) *core.ConfigMap) (*core.ConfigMap, error) {
	curJson, err := json.Marshal(cur)
	if err != nil {
		return nil, err
	}
	modJson, err := json.Marshal(transform(cur))
	if err != nil {
		return nil, err
	}
	patch, err := strategicpatch.CreateTwoWayMergePatch(curJson, modJson, apps.Deployment{})
	if err != nil {
		return nil, err
	}
	if len(patch) == 0 || string(patch) == "{}" {
		return cur, nil
	}
	return c.CoreV1().ConfigMaps(cur.Namespace).Patch(cur.Name, ktypes.StrategicMergePatchType, patch)
}

func initializerForResources(resources []string) v1alpha1.InitializerConfiguration {
	return v1alpha1.InitializerConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-initializer-config",
			Labels: map[string]string{
				"app": "my-app",
			},
		},
		Initializers: []v1alpha1.Initializer{
			{
				Name: initializerName,
				Rules: []v1alpha1.Rule{
					{
						APIGroups:   []string{"*"},
						APIVersions: []string{"*"},
						Resources:   resources,
					},
				},
			},
		},
	}
}

// Enable initializer:
// minikube start --extra-config=apiserver.Admission.PluginNames="Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,ResourceQuota"
