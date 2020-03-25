package k8s

import (
	"context"
	"github.com/keikoproj/manager/api/custom/v1alpha1"
	"github.com/keikoproj/manager/internal/utils"
	"github.com/keikoproj/manager/pkg/log"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/retry"
)

var (
	clusterGVR   = schema.GroupVersionResource{Group: "manager.keikoproj.io", Version: "v1alpha1", Resource: "cluster"}
	namespaceGVR = schema.GroupVersionResource{Group: "manager.keikoproj.io", Version: "v1alpha1", Resource: "namespace"}
)

//CreateOrUpdateClusterCR creates cluster custom resource
func (c *Client) CreateOrUpdateClusterCR(ctx context.Context, cr *v1alpha1.Cluster) error {
	log := log.Logger(ctx, "pkg.k8s", "resources", "CreateNamespace")
	_, err := c.CustomClient().CustomV1alpha1().Clusters(cr.Spec.Name).Create(cr)
	if err != nil {
		if !apierr.IsAlreadyExists(err) {
			log.Error(err, "unable to create cluster cr in the target namespace", "name", utils.SanitizeName(cr.Spec.Name))
			return err
		}
		//Modify the get response and retry update until no conflicts

		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			// Get the present CR to bump up the resource version
			resp, err := c.CustomClient().CustomV1alpha1().Clusters(cr.Spec.Name).Get(cr.Spec.Name, metav1.GetOptions{})
			if err != nil {
				log.Error(err, "unable to update cluster cr in the target namespace", "name", utils.SanitizeName(cr.Spec.Name))
				return err
			}

			resp.Spec = cr.Spec
			_, err = c.CustomClient().CustomV1alpha1().Clusters(cr.Spec.Name).Update(resp)
			return err
		})
		if retryErr != nil {
			log.Error(retryErr, "unable to update the cluster CR")
			return retryErr
		}
	}
	log.Info("Successfully cluster CR created/updated")
	return nil
}
