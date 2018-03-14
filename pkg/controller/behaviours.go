package controller

import (
	//"github.com/golang/glog"
	"fmt"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/apimachinery/pkg/api/errors"
	//"k8s.io/client-go/scale/scheme/appsv1beta2"
	"k8s.io/apimachinery/pkg/runtime/schema"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	somethingcontrollerv1alpha1 "sample-extension-apiserver/apis/somethingcontroller/v1alpha1"
)

// handleObject will take any resource implementing metav1.Object and attempt
// to find the Something resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that Something resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object, invalid type\n"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type\n"))
			return
		}
		fmt.Println(" Recovered deleted object '%s' from tombstone\n", object.GetName())
	}
	fmt.Printf("Processing object: %s\n", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a Something, we should not do anything more
		// with it.
		if ownerRef.Kind != "Something" {
			return
		}

		something, err := c.somethingsLister.Somethings(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			fmt.Println("ignoring orphaned object '%s' of something '%s'", object.GetSelfLink(), ownerRef.Name)
			return
		}

		if key, err := cache.MetaNamespaceKeyFunc(something); err == nil {
			c.somethingsQueue.Add(key)
		} else {
			runtime.HandleError(err)
		}

		return
	}
}

// Run will start the controller.
// StopCh channel is used to send interrupt signal to stop it.
func (c *Controller) Run(threadiness int, stopCh chan struct{}) error {
	// don't let panics crash the process
	defer runtime.HandleCrash()
	// make sure the work queue is shutdown which will trigger workers to end
	//defer c.deploymentsQueue.ShutDown()
	defer c.somethingsQueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	fmt.Println("Starting Something controller")

	// wait for the caches to synchronize before starting the worker
	fmt.Println("Waiting for informer caches to sync")
	if !cache.WaitForCacheSync(stopCh, c.somethingsInformer.HasSynced, c.deploymentsInformer.HasSynced) {
		fmt.Println("Timed out waiting for caches to sync")
		return fmt.Errorf("Timed out waiting for caches to sync")
	}

	fmt.Println("Starting workers")
	// Launch two workers to process Something resources
	// runWorker will loop until "something bad" happens.  The .Until will
	// then rekick the worker after one second
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runSomethingWorker, time.Second, stopCh)
		//go wait.Until(c.runDeploymentWorker, time.Second, stopCh)
	}

	fmt.Println("Started workers")
	<-stopCh
	fmt.Println("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runSomethingWorker() {
	for c.processNextSomethingWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextSomethingWorkItem() bool {
	obj, shutdown := c.somethingsQueue.Get()

	if shutdown {
		return false
	}

	// you always have to indicate to the queue that you've completed a piece of
	// work
	defer c.somethingsQueue.Done(obj)

	// do your work on the key.
	err := c.somethingSyncHandler(obj.(string))

	if err == nil {
		// No error, tell the queue to stop tracking history
		c.somethingsQueue.Forget(obj)
	} else if c.somethingsQueue.NumRequeues(obj) < 10 {
		fmt.Println("Error processing %s (will retry): %v", obj, err)
		// requeue the item to work on later
		c.somethingsQueue.AddRateLimited(obj)
	} else {
		// err != nil and too many retries
		fmt.Println("Error processing %s (giving up): %v", obj, err)
		c.somethingsQueue.Forget(obj)
		runtime.HandleError(err)
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Something resource
// with the current status of the resource.
func (c *Controller) somethingSyncHandler(key string) error {
	fmt.Println("handling the something resource named \"something-exmp\"...")
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s\n", key))
		return nil
	}
	fmt.Printf("key: %s, Namespace: %s, name: %s\n", key, namespace, name)

	// Get the Something resource with this namespace/name
	something, err := c.somethingsLister.Somethings(namespace).Get(name)
	if err != nil {
		// The Something resource may no longer exist, in which case we stop processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("something '%s' in work queue (somethingsQueue) no longer exists\n", key))
			return nil
		}
		fmt.Println("err in getting something resource is ", err)

		return err
	}
	fmt.Println(something.Namespace, "/", something.Name)

	//// initializer's tasks
	//if init_triggered, err := c.initialize(something); init_triggered {
	//	fmt.Println("occuring errror in \"initialize()\" method is", err)
	//	return err
	//}
	//
	//// finalizer's tasks
	//if fin_triggered, err := c.finalize(something); fin_triggered {
	//	fmt.Println("occuring errror in \"finalize()\" method is", err)
	//	return err
	//}

	deploymentName := something.Spec.DeploymentName
	if deploymentName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		runtime.HandleError(fmt.Errorf("%s: deployment name must be specified\n", key))
		return nil
	}
	fmt.Printf("deployment Name: %s\n", deploymentName)

	// Get the deployment with the name specified in Something.spec
	deployment, err := c.deploymentsLister.Deployments(something.Namespace).Get(deploymentName)
	// If the resource doesn't exist, we'll create it
	if errors.IsNotFound(err) {
		deployment, err = c.kubeclientset.AppsV1beta2().Deployments(something.Namespace).Create(newDeployment(something))
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		fmt.Println("error in getting/creating deploy is:", err)
		return err
	}

	// If the Deployment is not controlled by this Something resource, we should log
	// a warning to the event recorder and ret
	if !metav1.IsControlledBy(deployment, something) {
		msg := fmt.Sprintf("Resource %q already exists and is not managed by Something\n", deployment.Name)
		//c.recorder.Event(Something, corev1.EventTypeWarning, ErrResourceExists, msg)
		return fmt.Errorf(msg)
	}
	fmt.Println(deployment.Name, "is controlled by",something.Name)

	// If the number of the replicas on the Something resource is specified, and the
	// number does not equal the current desired replicas on the Deployment, we
	// should update the Deployment resource.
	if something.Spec.Replicas != nil && *something.Spec.Replicas != *deployment.Spec.Replicas {
		fmt.Println("SomethingR: %d, deployR: %d", *something.Spec.Replicas, *deployment.Spec.Replicas)
		deployment, err = c.kubeclientset.AppsV1beta2().Deployments(something.Namespace).Update(newDeployment(something))
	}

	// If an error occurs during Update, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		fmt.Println("error occured in updating deployment", deployment.Name, "is", err)
		return err
	}
	fmt.Println("no error in updating deployment", deployment.Name)

	// Finally, we update the status block of the Something resource to reflect the
	// current state of the world
	err = c.updateSomethingStatus(something, deployment)
	if err != nil {
		fmt.Println("error occured in updating status of something", something.Name, "is", err)
		return err
	}
	fmt.Println(something.Name, "is updated")

	//c.recorder.Event(foo, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) updateSomethingStatus(
		something *somethingcontrollerv1alpha1.Something,
		deployment *appsv1beta2.Deployment) error {
	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	somethingCopy := something.DeepCopy()
	somethingCopy.Status.AvailableReplicas = deployment.Status.AvailableReplicas
	// Until #38113 is merged, we must use Update instead of UpdateStatus to
	// update the Status block of the Foo resource. UpdateStatus will not
	// allow changes to the Spec of the resource, which is ideal for ensuring
	// nothing other than resource status has been updated.
	_, err := c.clientset.SomethingcontrollerV1alpha1().Somethings(something.Namespace).Update(somethingCopy)
	return err
}

// newDeployment creates a new Deployment for a Something resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Something resource that 'owns' it.
func newDeployment(something *somethingcontrollerv1alpha1.Something) *appsv1beta2.Deployment {
	labels := map[string]string{
		"app":        "book-server",
		"controller": something.Name,
	}
	return &appsv1beta2.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      something.Spec.DeploymentName,
			Namespace: something.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(something, schema.GroupVersionKind{
					Group:   somethingcontrollerv1alpha1.SchemeGroupVersion.Group,
					Version: somethingcontrollerv1alpha1.SchemeGroupVersion.Version,
					Kind:    "Something",
				}),
			},
		},
		Spec: appsv1beta2.DeploymentSpec{
			Replicas: something.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "book-server",
							Image: "shudipta/book_server:v1",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 10000,
								},
							},
						},
					},
				},
			},
		},
	}
}
