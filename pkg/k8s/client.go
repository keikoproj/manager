package k8s

import (
	"fmt"
	"github.com/keikoproj/manager/api/custom/v1alpha1"
	"github.com/keikoproj/manager/pkg/k8s/customclient"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Client struct {
	cl            kubernetes.Interface
	customCl      *customclient.Clientset
	runtimeClient client.Client
}

//NewK8sSelfClientDoOrDie gets the new k8s go client
func NewK8sSelfClientDoOrDie() *Client {
	config, err := rest.InClusterConfig()
	if err != nil {
		fmt.Println("THIS IS LOCAL")
		// Do i need to panic here?
		//How do i test this from local?
		//Lets get it from local config file
		config, err = clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	}
	cl, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	custom, err := customclient.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	//This is used for custom resources
	//https://godoc.org/sigs.k8s.io/controller-runtime/pkg/client#New
	//Lets make sure we add all our custom types to the scheme
	scheme, err := v1alpha1.SchemeBuilder.Register(&v1alpha1.ManagedNamespace{}, &v1alpha1.ManagedNamespaceList{}).Build()
	if err != nil {
		panic(err)
	}
	dClient, err := client.New(config, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		panic(err)
	}

	k8sCl := &Client{
		cl:            cl,
		customCl:      custom,
		runtimeClient: dClient,
	}
	return k8sCl
}

//NewK8sManagedClusterClientDoOrDie creates a client for managed cluster or config passed
func NewK8sManagedClusterClientDoOrDie(config *rest.Config) *Client {
	cl, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	//This is used for custom resources
	//https://godoc.org/sigs.k8s.io/controller-runtime/pkg/client#New
	dClient, err := client.New(config, client.Options{})
	if err != nil {
		panic(err)
	}

	k8sCl := &Client{
		cl:            cl,
		runtimeClient: dClient,
	}

	return k8sCl
}

func (c *Client) ClientInterface() kubernetes.Interface {
	return c.cl
}

func (c *Client) CustomClient() *customclient.Clientset {
	return c.customCl
}
