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
func (c *Client) CreateNamespace(ctx context.Context, ns *v1.Namespace) error {
	log := log.Logger(ctx, "pkg.k8s", "resources", "CreateNamespace")
	// Create the namespace
	resp, err := c.cl.CoreV1().Namespaces().Create(ns)
	if err != nil {
		if !apierr.IsAlreadyExists(err) {
			msg := fmt.Sprintf("unable to create the namespace %s", ns.Name)
			log.Error(err, msg)
			return errors.New(msg)
		}
		log.Info("Namespace already exists", "name", ns.Name)
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
func (c *Client) getNamespace(ctx context.Context, name string) error {
	log := log.Logger(ctx, "pkg.k8s", "resources", "CreateNamespace")
	// Create the namespace
	resp, err := c.cl.CoreV1().Namespaces().Get(name, metav1.GetOptions{})
	if err != nil {
		log.Error(err, "unable to get the namespace details")
		return nil
	}

	log.Info("Successfully created namespace", "name", resp.Name)
	return nil
}
