package controller

import (
	"fmt"
	"log"
	"time"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"

	clientset "k8s-extension-apiserver/client/clientset/versioned"
	fooscheme "k8s-extension-apiserver/client/clientset/versioned/scheme"
	informers "k8s-extension-apiserver/client/informers/externalversions"
	listers "k8s-extension-apiserver/client/listers/foocontroller/v1alpha1"
)

const (
	MaxRequeues           = 3
	SuccessSynced         = "Synced"
	MessageResourceSynced = "Foo synced successfully"
	EventComponent        = "foo-controller"
	ResyncPeriod          = time.Second * 30
)

// Controller is the controller implementation for Foo resources
type Controller struct {
	kubeClient         kubernetes.Interface
	fooClient          clientset.Interface
	fooInformerFactory informers.SharedInformerFactory
	foosLister         listers.FooLister
	foosSynced         cache.InformerSynced
	workqueue          workqueue.RateLimitingInterface
	recorder           record.EventRecorder
}

func NewController(kubeClient kubernetes.Interface, fooClient clientset.Interface) *Controller {
	// Create event broadcaster
	fooscheme.AddToScheme(scheme.Scheme)
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(glog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: EventComponent})

	fooInformerFactory := informers.NewSharedInformerFactory(fooClient, ResyncPeriod)
	fooInformer := fooInformerFactory.Foocontroller().V1alpha1().Foos()

	controller := &Controller{
		kubeClient:         kubeClient,
		fooClient:          fooClient,
		fooInformerFactory: fooInformerFactory,
		foosLister:         fooInformer.Lister(),
		foosSynced:         fooInformer.Informer().HasSynced,
		workqueue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Foos"),
		recorder:           recorder,
	}

	fooInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueFoo,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueFoo(new)
		},
		DeleteFunc: controller.enqueueFoo,
	})

	return controller
}

func (c *Controller) enqueueFoo(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

func (c *Controller) Run(stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	c.fooInformerFactory.Start(stopCh)
	if ok := cache.WaitForCacheSync(stopCh, c.foosSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	go wait.Until(c.runWorker, time.Second, stopCh)
	log.Println("Controller started")
	<-stopCh
	glog.Info("Shutting down controller")

	return nil
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

func (c *Controller) processNextItem() bool {
	key, quit := c.workqueue.Get()
	if quit {
		return false
	}

	defer c.workqueue.Done(key)

	if err := c.syncHandler(key.(string)); err == nil {
		c.workqueue.Forget(key)
	} else if c.workqueue.NumRequeues(key) > MaxRequeues {
		runtime.HandleError(err)
		c.workqueue.Forget(key)
	} else {
		c.workqueue.AddRateLimited(key)
	}

	return true
}

func (c *Controller) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	foo, err := c.foosLister.Foos(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("foo '%s' in work queue no longer exists", key))
			return nil
		}
		return err
	}

	log.Printf("Sync/Add/Update for foo '%s'", key)

	// create/patch configMap specified in foo.Spec
	c.recorder.Event(foo, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}
