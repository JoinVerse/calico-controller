package main

import (
	"time"

	"context"
	"flag"

	k8sclient "github.com/JoinVerse/calico-k8s-sync/client"
	"github.com/JoinVerse/calico-k8s-sync/controller"

	calicoclient "github.com/projectcalico/libcalico-go/lib/client"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "", "Path to a kube config. Only required if out-of-cluster.")
	flag.Parse()

	// Create the client config. Use kubeconfig if given, otherwise assume in-cluster.
	k8sConfig, err := buildConfig(*kubeconfig)
	if err != nil {
		panic(err)
	}

	// make a new config for our extension's API group, using the first config as a baseline
	k8sClient, k8sScheme, err := k8sclient.NewClient(k8sConfig)
	if err != nil {
		panic(err)
	}

	// NewFromEnv() creates a new client and defaults to access an etcd backend datastore at
	// http://127.0.0.1:2379.  For alternative backend access details, set the appropriate
	// ENV variables specified in the CalicoAPIConfigSpec structure.
	calicoClient, err := calicoclient.NewFromEnv()
	if err != nil {
		panic(err)
	}

	// start a controller on instances of our custom resource
	controller := controller.CalicoPolicyController{
		K8sClient:    k8sClient,
		K8sScheme:    k8sScheme,
		CalicoClient: calicoClient,
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	go controller.Run(ctx)
	time.Sleep(time.Second * 10000) // TODO LOL
}

func buildConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}
