package common

import (
	"context"
	"github.com/keikoproj/manager/internal/utils"
	"github.com/keikoproj/manager/pkg/k8s"
	"github.com/keikoproj/manager/pkg/log"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

	managerv1alpha1 "github.com/keikoproj/manager/api/custom/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
)

type StatusUpdatePredicate struct {
	predicate.Funcs
}

// Update implements default UpdateEvent filter for validating generation change
func (StatusUpdatePredicate) Update(e event.UpdateEvent) bool {
	log := log.Logger(context.Background(), "controllers.status", "status", "Update")
	if e.MetaOld == nil {
		log.Error(nil, "Update event has no old metadata", "event", e)
		return false
	}
	if e.ObjectOld == nil {
		log.Error(nil, "Update event has no old runtime object to update", "event", e)
		return false
	}
	if e.ObjectNew == nil {
		log.Error(nil, "Update event has no new runtime object for update", "event", e)
		return false
	}
	if e.MetaNew == nil {
		log.Error(nil, "Update event has no new metadata", "event", e)
		return false
	}
	//Better way to do it is to get GVK from ObjectKind but Kind is dropped during decode.
	//For more details, check the status of the issue here
	//https://github.com/kubernetes/kubernetes/issues/80609

	// Try to type caste to cluster first if it doesn't work move to namespace type casting
	if oldClusterObj, ok := e.ObjectOld.(*managerv1alpha1.Cluster); ok {
		newClusterObj := e.ObjectNew.(*managerv1alpha1.Cluster)
		if oldClusterObj.Status != newClusterObj.Status {
			return false
		}
	} else if oldNamespaceObj, ok := e.ObjectOld.(*managerv1alpha1.ManagedNamespace); ok {
		newNamespaceObj := e.ObjectNew.(*managerv1alpha1.ManagedNamespace)
		if oldNamespaceObj.Status != newNamespaceObj.Status {
			return false
		}
	}

	return true
}

// Client is a manager client to get the common stuff for all the controllers
type Client struct {
	client.Client
	K8sSelfClient *k8s.Client
	Recorder      record.EventRecorder
}

//UpdateMeta function updates the metadata (mostly finalizers in this case)
//This function accepts runtime.Object which can be either cluster type or namespace type
func (r *Client) UpdateMeta(ctx context.Context, object runtime.Object) {
	log := log.Logger(ctx, "controllers.common", "CommonClient", "UpdateMeta")
	if err := r.Update(ctx, object); err != nil {
		log.Error(err, "Unable to update object metadata (finalizer)")
		panic(err)
	}
}

//UpdateStatus function updates the status based on the process step
func (r *Client) UpdateStatus(ctx context.Context, obj runtime.Object, state managerv1alpha1.State, requeueTime ...float64) (ctrl.Result, error) {
	log := log.Logger(ctx, "controllers.common", "common", "UpdateStatus")

	if err := r.Status().Update(ctx, obj); err != nil {
		log.Error(err, "Unable to update status", "status", state)
		r.Recorder.Event(obj, v1.EventTypeWarning, string(managerv1alpha1.Error), "Unable to create/update status due to error "+err.Error())
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	if state != managerv1alpha1.Error {
		return ctrl.Result{}, nil
	}

	//if wait time is specified, requeue it after provided time
	if len(requeueTime) == 0 {
		requeueTime[0] = 0
	}

	log.Info("Requeue time", "time", requeueTime[0])
	return ctrl.Result{RequeueAfter: time.Duration(requeueTime[0]) * time.Millisecond}, nil
}

//ClusterConfig function returns cluster rest config for the managed cluster
func (r *Client) ClusterConfig(ctx context.Context, cluster *managerv1alpha1.Cluster) (*rest.Config, error) {
	log := log.Logger(ctx, "controllers.common", "ClusterConfig")

	secret, err := r.K8sSelfClient.GetK8sSecret(ctx, cluster.Spec.Config.BearerTokenSecret, cluster.ObjectMeta.Namespace)
	if err != nil {
		log.Error(err, "unable to retrieve the bearer token for the given cluster")
		return nil, err
	}
	cfg, err := utils.PrepareK8sRestConfigFromClusterCR(ctx, cluster, secret)
	if err != nil {
		log.Error(err, "unable to prepare the rest config for the target cluster", "cluster", cluster.Spec.Name)
		return nil, err
	}
	return cfg, nil
}

//ManagedClusterK8sClient returns the k8s client struct for the managed cluster
func (r *Client) ManagedClusterK8sClient(ctx context.Context, cluster *managerv1alpha1.Cluster) (*k8s.Client, error) {
	log := log.Logger(ctx, "controllers.common.common", "ManagedClusterK8sClient")
	config, err := r.ClusterConfig(ctx, cluster)
	if err != nil {
		log.Error(err, "unable to prepare the rest config for the target cluster", "cluster", cluster.Spec.Name)
		return nil, err
	}
	return k8s.NewK8sManagedClusterClientDoOrDie(config), nil
}
