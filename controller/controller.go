package controller

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	calicopolicyv1 "github.com/JoinVerse/calico-k8s-sync/apis/calicopolicy/v1"

	calicoapi "github.com/projectcalico/libcalico-go/lib/api"
	calicoclient "github.com/projectcalico/libcalico-go/lib/client"
)

type CalicoPolicyController struct {
	K8sClient    *rest.RESTClient
	K8sScheme    *runtime.Scheme
	CalicoClient *calicoclient.Client
}

// Run starts an CalicoPolicy resource controller
func (c *CalicoPolicyController) Run(ctx context.Context) error {
	fmt.Print("Watch CalicoPolicy objects\n")

	// Watch CalicoPolicy objects
	_, err := c.watchCalicoPolicies(ctx)
	if err != nil {
		fmt.Printf("Failed to register watch for CalicoPolicy resource: %v\n", err)
		return err
	}

	<-ctx.Done()
	return ctx.Err()
}

func (c *CalicoPolicyController) watchCalicoPolicies(ctx context.Context) (cache.Controller, error) {
	source := cache.NewListWatchFromClient(
		c.K8sClient,
		calicopolicyv1.CalicoPolicyResourcePlural,
		"",
		fields.Everything())

	store, controller := cache.NewInformer(
		source,

		// The object type.
		&calicopolicyv1.CalicoPolicy{},

		// resyncPeriod
		// Every resyncPeriod, all resources in the cache will retrigger events.
		// Set to 0 to disable the resync.
		time.Second*3,

		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.onAdd,
			UpdateFunc: c.onUpdate,
			DeleteFunc: c.onDelete,
		})

	fmt.Print(store.List())

	go controller.Run(ctx.Done())
	return controller, nil
}

func (c *CalicoPolicyController) onAdd(obj interface{}) {
	calicoPolicy := obj.(*calicopolicyv1.CalicoPolicy)
	fmt.Printf("[CONTROLLER] OnAdd %s\n", calicoPolicy.ObjectMeta.SelfLink)

	pol := &calicoapi.Policy{
		Metadata: calicoapi.PolicyMetadata{
			Name: calicoPolicy.ObjectMeta.Name,
		},
	}
	_, err := c.CalicoClient.Policies().Apply(pol)
	if err != nil {
		fmt.Printf("[CONTROLLER] Error applying calico policy: %v\n", err)
		return
	}
	fmt.Printf("[CONTROLLER] applied calico policy OK\n")

	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use scheme.Copy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance
	copyObj, err := c.K8sScheme.Copy(calicoPolicy)
	if err != nil {
		fmt.Printf("ERROR creating a deep copy of example object: %v\n", err)
		return
	}

	exampleCopy := copyObj.(*calicopolicyv1.CalicoPolicy)
	exampleCopy.Status = calicopolicyv1.CalicoPolicyStatus{
		State:   calicopolicyv1.CalicoPolicyStateProcessed,
		Message: "Successfully processed by controller",
	}

	err = c.K8sClient.Put().
		Name(calicoPolicy.ObjectMeta.Name).
		//		Namespace(example.ObjectMeta.Namespace).
		Resource(calicopolicyv1.CalicoPolicyResourcePlural).
		Body(exampleCopy).
		Do().
		Error()

	if err != nil {
		fmt.Printf("ERROR updating status: %v\n", err)
	} else {
		fmt.Printf("UPDATED status: %#v\n", exampleCopy)
	}
}

func (c *CalicoPolicyController) onUpdate(oldObj, newObj interface{}) {
	oldCalicoPolicy := oldObj.(*calicopolicyv1.CalicoPolicy)
	newCalicoPolicy := newObj.(*calicopolicyv1.CalicoPolicy)
	fmt.Printf("[CONTROLLER] OnUpdate oldObj: %s\n", oldCalicoPolicy.ObjectMeta.SelfLink)
	fmt.Printf("[CONTROLLER] OnUpdate newObj: %s\n", newCalicoPolicy.ObjectMeta.SelfLink)

	pol := &calicoapi.Policy{
		Metadata: calicoapi.PolicyMetadata{
			Name: newCalicoPolicy.ObjectMeta.Name,
		},
		Spec: newCalicoPolicy.Spec,
	}
	_, err := c.CalicoClient.Policies().Apply(pol)
	if err != nil {
		fmt.Printf("[CONTROLLER] Error applying calico policy: %v\n", err)
		return
	}
	fmt.Printf("[CONTROLLER] applied calico policy OK\n")

}

func (c *CalicoPolicyController) onDelete(obj interface{}) {
	calicoPolicy := obj.(*calicopolicyv1.CalicoPolicy)
	fmt.Printf("[CONTROLLER] OnDelete %s\n", calicoPolicy.ObjectMeta.SelfLink)

	err := c.CalicoClient.Policies().Delete(calicoapi.PolicyMetadata{Name: calicoPolicy.ObjectMeta.Name})
	if err != nil {
		fmt.Printf("[CONTROLLER] Error deleting calico policy: %v\n", err)
		return
	}
}
