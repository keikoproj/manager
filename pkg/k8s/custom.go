package k8s

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/keikoproj/manager/api/v1alpha1"
	"github.com/keikoproj/manager/pkg/grpc/proto/namespace"
	"github.com/keikoproj/manager/pkg/log"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	clusterGVR   = schema.GroupVersionResource{Group: "manager.keikoproj.io", Version: "v1alpha1", Resource: "cluster"}
	namespaceGVR = schema.GroupVersionResource{Group: "manager.keikoproj.io", Version: "v1alpha1", Resource: "namespace"}
)

//CreateOrUpdateClusterCR creates cluster custom resource
func (c *Client) CreateOrUpdateManagedCluster(ctx context.Context, cr *v1alpha1.Cluster, ns string) error {
	log := log.Logger(ctx, "pkg.k8s", "custom", "CreateOrUpdateManagedCluster")
	cr.SetNamespace(ns)
	cr.SetGroupVersionKind(cr.TypeMeta.GroupVersionKind())
	err := c.runtimeClient.Create(ctx, cr)
	if err != nil {
		if !apierr.IsAlreadyExists(err) {
			msg := fmt.Sprintf("unable to create the managed cluster")
			log.Error(err, msg)
			return errors.New(msg)
		}
		log.Info("managed cluster already exists. Trying to update")

		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {

			temp := v1alpha1.Cluster{}
			err = c.runtimeClient.Get(ctx, client.ObjectKey{
				Namespace: ns,
				Name:      cr.Name,
			}, &temp)

			if err != nil {
				log.Error(err, "unable to get the managed cluster")
				return err
			}
			rV := temp.GetResourceVersion()
			cr.SetResourceVersion(rV)

			err = c.runtimeClient.Update(ctx, cr)
			if err != nil {
				log.Error(err, "unable to update the managed cluster")
				return err
			}

			return nil
		})

		if retryErr != nil {
			log.Error(retryErr, "unable to update the cluster CR")
			return retryErr
		}
	}
	log.Info("Successfully created/updated managed cluster", "name", cr.Name)
	return nil
}

//CreateCustomResource creates a custom resource
func (c *Client) CreateOrUpdateCustomResource(ctx context.Context, cr *namespace.CustomResource, ns string) error {
	log := log.Logger(ctx, "pkg.k8s", "custom", "CreateCustomResource")

	jsonMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(cr.Manifest), &jsonMap)
	if err != nil {
		log.Error(err, "unable to unmarshal cr manifest to map[string]interface{}")
		return err
	}
	log.V(1).Info("unmarshalled manifest", "jsonMap", jsonMap)
	// Using a unstructured object.
	u := &unstructured.Unstructured{}
	u.Object = jsonMap
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   cr.GVK.Group,
		Kind:    cr.GVK.Kind,
		Version: cr.GVK.Version,
	})
	u.SetNamespace(ns)
	meta := jsonMap["metadata"].(map[string]interface{})
	name := meta["name"].(string)
	err = c.runtimeClient.Create(ctx, u)
	if err != nil {
		if !apierr.IsAlreadyExists(err) {
			msg := fmt.Sprintf("unable to create the custom resource")
			log.Error(err, msg)
			return errors.New(msg)
		}
		log.Info("custom resource already exists. Trying to update")

		err = c.runtimeClient.Get(ctx, client.ObjectKey{
			Namespace: ns,
			Name:      name,
		}, u)

		if err != nil {
			log.Error(err, "unable to get the custom resource")
			return err
		}
		rV := u.GetResourceVersion()
		u.SetUnstructuredContent(jsonMap)
		u.SetResourceVersion(rV)

		log.Info("custom resource ", "custom", u)
		err = c.runtimeClient.Update(ctx, u)
		if err != nil {
			log.Error(err, "unable to update the custom resource")
			return err
		}
	}
	log.Info("Successfully created custom resource", "name", name)
	return nil
}

//CreateOrUpdateManagedNamespace creates/updates managed namespace
func (c *Client) CreateOrUpdateManagedNamespace(ctx context.Context, cr *v1alpha1.ManagedNamespace, ns string) error {
	log := log.Logger(ctx, "pkg.k8s", "custom", "CreateOrUpdateManagedNamespace")
	cr.SetNamespace(ns)
	cr.SetGroupVersionKind(cr.TypeMeta.GroupVersionKind())
	err := c.runtimeClient.Create(ctx, cr)
	if err != nil {
		if !apierr.IsAlreadyExists(err) {
			msg := fmt.Sprintf("unable to create the managed namespace")
			log.Error(err, msg)
			return errors.New(msg)
		}
		log.Info("managed namespace already exists. Trying to update")
		temp := v1alpha1.ManagedNamespace{}
		err = c.runtimeClient.Get(ctx, client.ObjectKey{
			Namespace: ns,
			Name:      cr.Name,
		}, &temp)

		if err != nil {
			log.Error(err, "unable to get the managed namespace")
			return err
		}
		rV := temp.GetResourceVersion()
		cr.SetResourceVersion(rV)

		err = c.runtimeClient.Update(ctx, cr)
		if err != nil {
			log.Error(err, "unable to update the managed namespace")
			return err
		}
	}
	log.Info("Successfully created/updated managed namespace", "name", cr.Name)
	return nil
}

//DeleteManagedCluster deletes managed cluster with propagation policy foreground
func (c *Client) DeleteManagedCluster(ctx context.Context, cr *v1alpha1.Cluster, ns string) error {
	log := log.Logger(ctx, "pkg.k8s", "custom", "DeleteManagedCluster")

	cr.SetNamespace(ns)
	deletePolicy := metav1.DeletePropagationForeground
	dOptions := client.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}
	newDeleteOpts := &client.DeleteOptions{}
	dOptions.ApplyToDelete(newDeleteOpts)
	err := c.runtimeClient.Delete(ctx, cr, newDeleteOpts)
	if err != nil {
		log.Error(err, "unable to delete the managed cluster")
		return err
	}

	return nil
}
