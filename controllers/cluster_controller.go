/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"github.com/keikoproj/manager/internal/config"
	"github.com/keikoproj/manager/internal/utils"
	"github.com/keikoproj/manager/pkg/k8s"
	"github.com/keikoproj/manager/pkg/log"
	"github.com/pborman/uuid"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	managerv1alpha1 "github.com/keikoproj/manager/api/custom/v1alpha1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
)

const (
	finalizerName = "cluster.finalizers.manager.keikoproj.io"
	requestId     = "request_id"
	//2 minutes
	maxWaitTime = 120000
	//30 seconds
	errRequeueTime = 30000
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Log       logr.Logger
	K8sClient *k8s.Client
	Recorder  record.EventRecorder
}

// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=manager.keikoproj.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=manager.keikoproj.io,resources=clusters/status,verbs=get;update;patch
//Main responsibilities of the cluster controller should be
//1. Handling service account bearer token rotation
//2. Validation of certain namespaces(??)
func (r *ClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	ctx := context.WithValue(context.Background(), requestId, uuid.New())
	log := log.Logger(ctx, "controllers", "cluster_controller", "Reconcile")
	log.WithValues("cluster", req.NamespacedName)
	log.Info("Start of the request")
	//Get the resource
	var cluster managerv1alpha1.Cluster
	if err := r.Get(ctx, req.NamespacedName, &cluster); err != nil {
		return ctrl.Result{}, ignoreNotFound(err)
	}

	state := managerv1alpha1.Warning
	if cluster.Status.RetryCount > 3 {
		state = managerv1alpha1.Error
	}

	log.Info("received", "cluster_secret", cluster.Spec.Config.BearerTokenSecret)
	// Retrieve k8s secret
	// Get the "best" Bearer token
	// Get the ManagedCluster k8s client
	// Validate the connection
	// Update the status
	// Requeue it based on config map variable
	secret, err := r.K8sClient.GetK8sSecret(ctx, cluster.Spec.Config.BearerTokenSecret, cluster.ObjectMeta.Namespace)
	if err != nil {
		log.Error(err, "unable to retrieve the bearer token for the given cluster")
		desc := fmt.Sprintf("unable to retrieve the bearer token for the given cluster due to error %s", err.Error())
		r.Recorder.Event(&cluster, v1.EventTypeWarning, string(state), desc)
		return r.UpdateStatus(ctx, &cluster, managerv1alpha1.ClusterStatus{RetryCount: cluster.Status.RetryCount + 1, ErrorDescription: desc}, state, errRequeueTime)
	}
	cfg, err := utils.PrepareK8sRestConfigFromClusterCR(ctx, &cluster, secret)
	if err != nil {
		log.Error(err, "unable to prepare the rest config for the target cluster", "cluster", cluster.Spec.Name)
		desc := fmt.Sprintf("unable to prepare the rest config for the target cluster due to error %s", err.Error())
		r.Recorder.Event(&cluster, v1.EventTypeWarning, string(state), desc)
		return r.UpdateStatus(ctx, &cluster, managerv1alpha1.ClusterStatus{RetryCount: cluster.Status.RetryCount + 1, ErrorDescription: desc}, state, errRequeueTime)
	}
	// Ge the Managed cluster client
	resp, err := GetServerVersion(cfg)
	if err != nil {
		log.Error(err, "unable to get the server version", "cluster", cluster.Spec.Name)
		desc := fmt.Sprintf("Unable to get the server version due to error %s", err.Error())
		r.Recorder.Event(&cluster, v1.EventTypeWarning, string(state), desc)
		return r.UpdateStatus(ctx, &cluster, managerv1alpha1.ClusterStatus{RetryCount: cluster.Status.RetryCount + 1, ErrorDescription: desc}, state, errRequeueTime)
	}
	r.Recorder.Event(&cluster, v1.EventTypeNormal, string(managerv1alpha1.Ready), "Successfully validated the target cluster")
	r.UpdateStatus(ctx, &cluster, managerv1alpha1.ClusterStatus{RetryCount: 0, ErrorDescription: ""}, managerv1alpha1.Ready)
	log.Info("SUCCESSFUL", "version", resp)
	return ctrl.Result{RequeueAfter: time.Duration(config.Props.ClusterValidationFrequency()) * time.Second}, nil
}

type StatusUpdatePredicate struct {
	predicate.Funcs
}

// Update implements default UpdateEvent filter for validating generation change
func (StatusUpdatePredicate) Update(e event.UpdateEvent) bool {
	log := log.Logger(context.Background(), "controllers", "iamrole_controller", "HandleReconcile")
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
	oldObj := e.ObjectOld.(*managerv1alpha1.Cluster)
	newObj := e.ObjectNew.(*managerv1alpha1.Cluster)

	if oldObj.Status != newObj.Status {
		return false
	}
	return true
}

//UpdateStatus function updates the status based on the process step
func (r *ClusterReconciler) UpdateStatus(ctx context.Context, cluster *managerv1alpha1.Cluster, status managerv1alpha1.ClusterStatus, state managerv1alpha1.State, requeueTime ...float64) (ctrl.Result, error) {
	log := log.Logger(ctx, "controllers", "cluster_controller", "UpdateStatus")
	log.WithValues("cluster", cluster.ObjectMeta.Name)
	status.State = state
	cluster.Status = status
	if err := r.Status().Update(ctx, cluster); err != nil {
		log.Error(err, "Unable to update status", "status", state)
		r.Recorder.Event(cluster, v1.EventTypeWarning, string(managerv1alpha1.Error), "Unable to create/update status due to error "+err.Error())
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

//UpdateMeta function updates the metadata (mostly finalizers in this case)
func (r *ClusterReconciler) UpdateMeta(ctx context.Context, cluster *managerv1alpha1.Cluster) {
	log := log.Logger(ctx, "controllers", "cluster_controller", "UpdateMeta")
	log.WithValues("cluster", fmt.Sprintf("k8s-%s", cluster.ObjectMeta.Namespace))
	if err := r.Update(ctx, cluster); err != nil {
		log.Error(err, "Unable to update object metadata (finalizer)")
		panic(err)
	}
}

func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&managerv1alpha1.Cluster{}).
		WithEventFilter(StatusUpdatePredicate{}).
		Complete(r)
}

/*
We generally want to ignore (not requeue) NotFound errors, since we'll get a
reconciliation request once the object exists, and requeuing in the meantime
won't help.
*/
func ignoreNotFound(err error) error {
	if apierrs.IsNotFound(err) {
		return nil
	}
	return err
}

func GetServerVersion(config *rest.Config) (string, error) {
	client, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return "", err
	}

	v, err := client.ServerVersion()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s.%s", v.Major, v.Minor), nil
}
