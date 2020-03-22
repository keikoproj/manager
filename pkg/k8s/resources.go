package k8s

import (
	"context"
	"errors"
	"fmt"
	"github.com/keikoproj/manager/pkg/log"
	"k8s.io/api/core/v1"

	apierr "k8s.io/apimachinery/pkg/api/errors"
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
