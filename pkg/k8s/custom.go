package k8s

import (
	"context"
	"github.com/keikoproj/manager/api/custom/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	clusterGVR   = schema.GroupVersionResource{Group: "manager.keikoproj.io", Version: "v1alpha1", Resource: "cluster"}
	namespaceGVR = schema.GroupVersionResource{Group: "manager.keikoproj.io", Version: "v1alpha1", Resource: "namespace"}
)

//CreateClusterCR creates cluster custom resource
func (client *Client) CreateClusterCR(ctx context.Context, cr *v1alpha1.Cluster, ns string) error {

	//client.dCl.Resource(clusterGVR).Namespace(ns).Create(cr, metav1.CreateOptions{})

	return nil
}
