package controller

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	calicoapi "github.com/projectcalico/libcalico-go/lib/api"
	calicoclient "github.com/projectcalico/libcalico-go/lib/client"
	caliconet "github.com/projectcalico/libcalico-go/lib/net"
)

type HostEnpdointController struct {
	K8sClient    rest.Interface
	K8sScheme    *runtime.Scheme
	CalicoClient *calicoclient.Client
}

func (c *HostEnpdointController) Run(ctx context.Context) error {
	fmt.Printf("[HOST ENDPOINT CONTROLLER] Starting up...\n")

	source := cache.NewListWatchFromClient(
		c.K8sClient,
		"nodes",
		"",
		fields.Everything())

	_, controller := cache.NewInformer(
		source,

		// The object type.
		&v1.Node{},

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

	fmt.Printf("[HOST ENDPOINT CONTROLLER] Exited.\n")
	return nil
}

func (c *HostEnpdointController) applyNode(node *v1.Node) {
	ips := []caliconet.IP{}
	for _, ip := range node.Status.Addresses {
		if ip.Type == v1.NodeInternalIP {
			ips = append(ips, *caliconet.ParseIP(ip.Address))
		}
	}

	labels := make(map[string]string)
	if _, ok := node.ObjectMeta.Labels["node-role.kubernetes.io/master"]; ok {
		labels["k8s_node/role"] = "master"
	} else {
		labels["k8s_node/role"] = "node"
	}

	hep := &calicoapi.HostEndpoint{
		Metadata: calicoapi.HostEndpointMetadata{
			Name:   node.ObjectMeta.Name,
			Node:   node.ObjectMeta.Name,
			Labels: labels,
		},
		Spec: calicoapi.HostEndpointSpec{
			ExpectedIPs:   ips,
			InterfaceName: "ens4",
		},
	}
	_, err := c.CalicoClient.HostEndpoints().Apply(hep)
	if err != nil {
		fmt.Printf("[HOST ENDPOINT CONTROLLER] Error applying hostendpoint: %v\n", err)
		return
	}
	fmt.Printf("[HOST ENDPOINT CONTROLLER] applied hostendpoint OK\n")
}

func (c *HostEnpdointController) onAdd(obj interface{}) {
	node := obj.(*v1.Node)
	fmt.Printf("[HOST ENDPOINT CONTROLLER] OnAdd %s\n", node.ObjectMeta.SelfLink)

	c.applyNode(node)
}

func (c *HostEnpdointController) onUpdate(oldObj, newObj interface{}) {
	node := newObj.(*v1.Node)
	fmt.Printf("[HOST ENDPOINT CONTROLLER] OnUpdate: %s\n", node.ObjectMeta.SelfLink)
	c.applyNode(node)
}

func (c *HostEnpdointController) onDelete(obj interface{}) {
	node := obj.(*v1.Node)
	fmt.Printf("[HOST ENDPOINT CONTROLLER] OnDelete %s\n", node.ObjectMeta.SelfLink)

	err := c.CalicoClient.HostEndpoints().Delete(calicoapi.HostEndpointMetadata{Name: node.ObjectMeta.Name})
	if err != nil {
		fmt.Printf("[HOST ENDPOINT CONTROLLER] Error deleting hostendpoint: %v\n", err)
	}
	err = c.CalicoClient.Nodes().Delete(calicoapi.NodeMetadata{Name: node.ObjectMeta.Name})
	if err != nil {
		fmt.Printf("[HOST ENDPOINT CONTROLLER] Error deleting node: %v\n", err)
	}
}
