package k8s

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

type Client struct {
	cl  kubernetes.Interface
	dCl dynamic.Interface
}

//NewK8sSelfClientDoOrDie gets the new k8s go client
func NewK8sSelfClientDoOrDie() *Client {
	config, err := rest.InClusterConfig()
	if err != nil {
		// Do i need to panic here?
		//How do i test this from local?
		//Lets get it from local config file
		config, err = clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	dClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	cl := &Client{
		client,
		dClient,
	}
	return cl
}

//NewK8sManagedClusterClientDoOrDie creates a client for managed cluster or config passed
func NewK8sManagedClusterClientDoOrDie(client *kubernetes.Clientset) *Client {
	cl := &Client{
		cl: client,
	}

	return cl
}

func (c *Client) ClientInterface() kubernetes.Interface {
	return c.cl
}
