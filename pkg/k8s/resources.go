package k8s

import (
	"context"
	"errors"
	"fmt"
	"github.com/keikoproj/manager/pkg/log"
	"k8s.io/api/core/v1"

	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//CreateNamespace function creates a namespace in the control plan cluster
func (c *Client) CreateOrUpdateNamespace(ctx context.Context, ns *v1.Namespace) error {
	log := log.Logger(ctx, "pkg.k8s", "resources", "CreateNamespace")
	// Create the namespace
	resp, err := c.cl.CoreV1().Namespaces().Create(ns)
	if err != nil {
		if !apierr.IsAlreadyExists(err) {
			msg := fmt.Sprintf("unable to create the namespace %s", ns.Name)
			log.Error(err, msg)
			return errors.New(msg)
		}
		log.Info("Namespace already exists.. Trying to update", "name", ns.Name)
		resp, err = c.cl.CoreV1().Namespaces().Update(ns)
		if err != nil {
			msg := fmt.Sprintf("Failed to update namespace %s due to %v", ns.Name, err)
			log.Error(err, msg)
			return errors.New(msg)
		}
		return nil
	}

	log.Info("Successfully created namespace", "name", resp.Name)
	return nil
}

//DeleteNamespace function creates a namespace in the control plan cluster
func (c *Client) DeleteNamespace(ctx context.Context, name string) error {
	log := log.Logger(ctx, "pkg.k8s", "resources", "DeleteNamespace")
	// Delete the namespace
	err := c.cl.CoreV1().Namespaces().Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		if !apierr.IsNotFound(err) {
			msg := fmt.Sprintf("unable to delete the namespace %s", name)
			log.Error(err, msg)
			return errors.New(msg)
		}
		log.Info("Namespace doesn't exist anymore", "name", name)
		return nil
	}

	log.Info("Successfully deleted namespace", "name", name)
	return nil
}

//CreateNamespace function creates a namespace in the control plan cluster
func (c *Client) GetNamespace(ctx context.Context, name string) error {
	log := log.Logger(ctx, "pkg.k8s", "resources", "GetNamespace")
	// Create the namespace
	resp, err := c.cl.CoreV1().Namespaces().Get(name, metav1.GetOptions{})
	if err != nil {
		log.Error(err, "unable to get the namespace details")
		return err
	}

	log.Info("Successfully created namespace", "name", resp.Name)
	return nil
}

//CreateResourceQuota function creates resource quota for a specified namespace
func (c *Client) CreateOrUpdateResourceQuota(ctx context.Context, quota *v1.ResourceQuota, ns string) error {
	log := log.Logger(ctx, "pkg.k8s", "resources", "CreateResourceQuota")
	log = log.WithValues("namespace", ns, "quotaName", quota.Name)

	_, err := c.cl.CoreV1().ResourceQuotas(ns).Create(quota)
	if err != nil {
		if !apierr.IsAlreadyExists(err) {
			msg := fmt.Sprintf("unable to create the resource quota %s", quota.Name)
			log.Error(err, msg)
			return errors.New(msg)
		}
		log.Info("Resource quota already exists. Trying to update")
		_, err := c.cl.CoreV1().ResourceQuotas(ns).Update(quota)
		if err != nil {
			msg := fmt.Sprintf("Failed to update resource quota %s due to %v", quota.Name, err)
			log.Error(err, msg)
			return errors.New(msg)
		}
		return nil
	}
	log.Info("successfully created resource quota")
	return nil
}
