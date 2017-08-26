package controller

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	calicov1 "github.com/JoinVerse/calico-controller/apis/calico/v1"

	calicoapi "github.com/projectcalico/libcalico-go/lib/api"
	calicoclient "github.com/projectcalico/libcalico-go/lib/client"
)

type PolicyController struct {
	K8sClient    rest.Interface
	K8sScheme    *runtime.Scheme
	CalicoClient *calicoclient.Client
}

// Run starts an CalicoPolicy resource controller
func (c *PolicyController) Run(ctx context.Context) error {
	fmt.Printf("[POLICY CONTROLLER] Starting up...\n")

	source := cache.NewListWatchFromClient(
		c.K8sClient,
		calicov1.CalicoPolicyResourcePlural,
		"",
		fields.Everything())

	_, controller := cache.NewInformer(
		source,

		// The object type.
		&calicov1.CalicoPolicy{},

		// resyncPeriod
		// Every resyncPeriod, all resources in the cache will retrigger events.
		// Set to 0 to disable the resync.
		time.Second*300,

		// Your custom resource event handlers.
		cache.ResourceEventHandlerFuncs{
			AddFunc:    c.onAdd,
			UpdateFunc: c.onUpdate,
			DeleteFunc: c.onDelete,
		})

	controller.Run(ctx.Done())
	fmt.Printf("[POLICY CONTROLLER] Exited.\n")

	return nil
}

func (c *PolicyController) applyPolicy(polObj *calicov1.CalicoPolicy) {
	pol := &calicoapi.Policy{
		Metadata: calicoapi.PolicyMetadata{
			Name: polObj.ObjectMeta.Name,
		},
		Spec: polObj.Spec,
	}
	_, err := c.CalicoClient.Policies().Apply(pol)
	if err != nil {
		fmt.Printf("[POLICY CONTROLLER] Error applying calico policy: %v\n", err)
		return
	}
}

func (c *PolicyController) onAdd(obj interface{}) {
	calicoPolicy := obj.(*calicov1.CalicoPolicy)
	fmt.Printf("[POLICY CONTROLLER] OnAdd %s\n", calicoPolicy.ObjectMeta.SelfLink)

	c.applyPolicy(calicoPolicy)
}

func (c *PolicyController) onUpdate(oldObj, newObj interface{}) {
	calicoPolicy := newObj.(*calicov1.CalicoPolicy)
	fmt.Printf("[POLICY CONTROLLER] OnUpdate: %s\n", calicoPolicy.ObjectMeta.SelfLink)
	c.applyPolicy(calicoPolicy)
}

func (c *PolicyController) onDelete(obj interface{}) {
	calicoPolicy := obj.(*calicov1.CalicoPolicy)
	fmt.Printf("[POLICY CONTROLLER] OnDelete %s\n", calicoPolicy.ObjectMeta.SelfLink)

	err := c.CalicoClient.Policies().Delete(calicoapi.PolicyMetadata{Name: calicoPolicy.ObjectMeta.Name})
	if err != nil {
		fmt.Printf("[POLICY CONTROLLER] Error deleting calico policy: %v\n", err)
		return
	}
}
