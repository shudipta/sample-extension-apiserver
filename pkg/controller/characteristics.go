package controller

import (
	"k8s.io/client-go/kubernetes"
	clientset "sample-extension-apiserver/client/clientset/versioned"
	listers "sample-extension-apiserver/client/listers/somethingcontroller/v1alpha1"
	kubelisters "k8s.io/client-go/listers/apps/v1beta2"
	"k8s.io/client-go/util/workqueue"
	//"k8s.io/client-go/tools/record"
	"k8s.io/client-go/tools/cache"
	//"github.com/golang/glog"
	//"k8s.io/client-go/scale/scheme/appsv1beta2"
	kubeinformers "k8s.io/client-go/informers"
	informers "sample-extension-apiserver/client/informers/externalversions"
	//"k8s.io/client-go/scale/scheme/appsv1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	somethingcontrollerv1alpha1 "sample-extension-apiserver/apis/somethingcontroller/v1alpha1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"fmt"
)

// Controller implementation for Something resources
type Controller struct {
	kubeclientset		kubernetes.Interface
	clientset			clientset.Interface

	deploymentsLister	kubelisters.DeploymentLister
	somethingsLister	listers.SomethingLister

	//deploymentsQueue	workqueue.RateLimitingInterface
	somethingsQueue		workqueue.RateLimitingInterface

	deploymentsInformer	cache.SharedIndexInformer
	somethingsInformer	cache.SharedIndexInformer
	//
	//deploymentsSynced	cache.InformerSynced
	//somethingsSynced	cache.InformerSynced

	//kubType				string
	//recorder			record.EventRecorder
}

// NewController returns a new sample-crd-controller
func NewController(
	kubeclientset kubernetes.Interface,
	clientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	sampleInformerFactory informers.SharedInformerFactory) *Controller {

	// obtain references to shared index informers for the Deployment and Something CRD
	// types.
	deploymentInformer := kubeInformerFactory.Apps().V1beta2().Deployments()
	somethingInformer := sampleInformerFactory.Somethingcontroller().V1alpha1().Somethings()

	controller := &Controller{
		kubeclientset:		kubeclientset,
		clientset:			clientset,

		deploymentsLister:	deploymentInformer.Lister(),
		somethingsLister:	somethingInformer.Lister(),

		deploymentsInformer:deploymentInformer.Informer(),
		somethingsInformer:	somethingInformer.Informer(),

		//deploymentsQueue:	workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Deployments"),
		somethingsQueue:	workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Somethings"),
		//recorder:          recorder,
	}

	fmt.Println("Setting up event handlers")
	// Set up an event handler for when Something resources change
	controller.somethingsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		//AddFunc: controller.handleObject,
		AddFunc: func(obj interface{}) {
			if key, err := cache.MetaNamespaceKeyFunc(obj); err == nil {
				controller.somethingsQueue.Add(key)
			} else {
				runtime.HandleError(err)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			newSomething := new.(*somethingcontrollerv1alpha1.Something)
			oldSomething := old.(*somethingcontrollerv1alpha1.Something)

			if oldSomething.ResourceVersion == newSomething.ResourceVersion {
				// Periodic resync will send update events for all known Somethings.
				// Two different versions of the same Somthing will always have different RVs.
				return
			} else {
				if key, err := cache.MetaNamespaceKeyFunc(new); err == nil {
					controller.somethingsQueue.Add(key)
				}
			}
			//controller.handleObject(new)
		},
		//DeleteFunc: controller.handleObject,
		DeleteFunc: func(obj interface{}) {
			if key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj); err == nil {
				controller.somethingsQueue.Add(key)
			}
		},
	})
	// Set up an event handler for when Deployment resources change. This
	// handler will lookup the owner of the given Deployment, and if it is
	// owned by a Something resource will enqueue that Something resource for
	// processing. This way, we don't need to implement custom logic for
	// handling Deployment resources. More info on this pattern:
	// https://github.com/kubernetes/community/blob/8cafef897a22026d42f5e5bb3f104febe7e29830/contributors/devel/controllers.md
	controller.deploymentsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		//AddFunc: func(obj interface{}) {
		//	if key, err := cache.MetaNamespaceKeyFunc(obj); err == nil {
		//		controller.deploymentsQueue.Add(key)
		//	}
		//},
		UpdateFunc: func(old, new interface{}) {
			newDeploy := new.(*appsv1beta2.Deployment)
			oldDeploy := old.(*appsv1beta2.Deployment)

			if newDeploy.ResourceVersion == oldDeploy.ResourceVersion {
				// Periodic resync will send update events for all known Deployments.
				// Two different versions of the same Deployment will always have different RVs.
				return
			}
			//else {
			//	if key, err := cache.MetaNamespaceKeyFunc(new); err == nil {
			//		controller.deploymentsQueue.Add(key)
			//	}
			//}
			controller.handleObject(new)
		},
		DeleteFunc: controller.handleObject,
		//DeleteFunc: func(obj interface{}) {
		//	if key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj); err == nil {
		//		controller.deploymentsQueue.Add(key)
		//	}
		//},
	})

	return controller
}
