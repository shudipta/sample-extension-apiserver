package controller

import (
	somethingcontrollerv1alpha1 "sample-extension-apiserver/apis/somethingcontroller/v1alpha1"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var initializerName = 	"fun.samplecrdcontroller.crd.com"
var finalizerName =		"endfun.samplecrdcontroller.crd.com"

func (c *Controller)initialize(something *somethingcontrollerv1alpha1.Something) (trigger bool, err error) {
	trigger = false
	err = nil
	if something.ObjectMeta.GetInitializers() != nil {
		trigger = true
		pendingInitializers := something.ObjectMeta.GetInitializers().Pending

		if initializerName == pendingInitializers[0].Name {
			fmt.Printf("\n>>>>>>> Initializing \"something\": \"%s\"...\n", something.Name)
			fmt.Printf("\n>>>>>>> Initializer name is %s\n", initializerName)
			fmt.Printf("\n>>>>>>> And the fun is 2 + 2 = %d", 2 + 2)

			somethingCopy := something.DeepCopy()
			//adding finalizer
			ch := false
			for _, name := range somethingCopy.ObjectMeta.Finalizers {
				if name == finalizerName {
					ch = true
					break
				}
			}
			if !ch {
				somethingCopy.ObjectMeta.Finalizers = append(somethingCopy.ObjectMeta.Finalizers, finalizerName)
			}
			fmt.Println("finalizer is added")

			// removing self from initializer
			fmt.Printf("\n>>>>>>> Removing initializer \"%s\"...", initializerName)
			if len(pendingInitializers) == 1 {
				somethingCopy.ObjectMeta.Initializers = nil
			} else {
				somethingCopy.ObjectMeta.Initializers.Pending = append(pendingInitializers[:0], pendingInitializers[1:]...)
			}
			_, err = c.clientset.SomethingcontrollerV1alpha1().Somethings(something.Namespace).Update(somethingCopy)

			return trigger, err
		}

		return trigger, err
	}

	return trigger, err
}

func (c *Controller) finalize(something *somethingcontrollerv1alpha1.Something) (trigger bool, err error) {
	trigger = false
	err = nil
	if(something.DeletionTimestamp != nil) {
		trigger = true

		got := false
		for _, name := range something.Finalizers {
			if name == finalizerName {
				got = true
				break
			}
		}

		if got {

			// deleting deployments
			labels := map[string]string{
				"app":        "book-server",
				"controller": something.Name,
			}
			for key, value := range labels {
				deployList, dlErr := c.kubeclientset.AppsV1beta2().Deployments(something.Namespace).List(
					metav1.ListOptions{LabelSelector: key + "=" + value},
				)
				if dlErr != nil {
					err = dlErr
					fmt.Println("failed deleting deployment")
					return trigger, err
				}

				for _, deploy := range deployList.Items {
					delErr := c.kubeclientset.AppsV1beta2().Deployments(something.Namespace).Delete(
						deploy.GetName(),
						&metav1.DeleteOptions{},
					)
					if delErr != nil {
						err = delErr
						fmt.Println("failed deleting deployment")
						return trigger, err
					}
				}
			}

			fmt.Printf("\n>>>>>>> finalizing \"something\": \"%s\"...\n", something.Name)
			fmt.Printf("\n>>>>>>> finalizer name is %s\n", finalizerName)
			fmt.Printf("\n>>>>>>> And the fun is ending with 2 * 2 = %d", 2 * 2)

			// removing finalizer
			somethingCopy := something.DeepCopy()
			r := somethingCopy.ObjectMeta.Finalizers[:0]
			for _, name := range somethingCopy.ObjectMeta.Finalizers {
				if name != finalizerName {
					r = append(r, name)
				}
			}
			somethingCopy.ObjectMeta.Finalizers = r
			_, upErr := c.clientset.SomethingcontrollerV1alpha1().Somethings(something.Namespace).Update(somethingCopy)
			if upErr != nil {
				err = upErr
				fmt.Println("error in removing finalizer is ", upErr)
				return trigger, err
			}
		}
	}

	return trigger, err
}
