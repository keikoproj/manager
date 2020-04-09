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
	common2 "github.com/keikoproj/manager/controllers/common"
	"github.com/keikoproj/manager/internal/config"
	"github.com/keikoproj/manager/internal/config/common"
	"github.com/keikoproj/manager/internal/utils"
	"github.com/keikoproj/manager/pkg/k8s"
	"github.com/keikoproj/manager/pkg/log"
	"github.com/pborman/uuid"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"time"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	managerv1alpha1 "github.com/keikoproj/manager/api/custom/v1alpha1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
)

const (
	clusterFinalizerName = "cluster.finalizers.manager.keikoproj.io"
	requestId            = "request_id"
	//2 minutes
	maxWaitTime = 120000
	//30 seconds
	errRequeueTime = 300000
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Log           logr.Logger
	Scheme        *runtime.Scheme
	K8sSelfClient *k8s.Client
	Recorder      record.EventRecorder
}

// +kubebuilder:rbac:groups=core,resources=secrets,verbs=delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=get;list;watch;create
// +kubebuilder:rbac:groups=manager.keikoproj.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=manager.keikoproj.io,resources=clusters/status,verbs=get;update;patch

func (r *ClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {

	//Main responsibilities of the cluster controller should be
	//1. Handling service account bearer token rotation
	//2. Validation of certain namespaces(??)

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	ctx := context.WithValue(context.Background(), requestId, uuid.New())
	log := log.Logger(ctx, "controllers", "cluster_controller", "Reconcile")
	log = log.WithValues("cluster", req.NamespacedName)
	log.Info("Start of the request")
	commonClient := &common2.Client{Client: r.Client, Recorder: r.Recorder}
	//Get the resource
	var cluster managerv1alpha1.Cluster
	if err := r.Get(ctx, req.NamespacedName, &cluster); err != nil {
		return ctrl.Result{}, ignoreNotFound(err)
	}

	// Retrieve k8s secret
	// Get the "best" Bearer token
	// Get the ManagedCluster k8s client

	state := managerv1alpha1.Warning

	if cluster.Status.RetryCount > 3 {
		state = managerv1alpha1.Error
	}

	cfg, err := commonClient.ClusterConfig(ctx, &cluster)
	if err != nil {
		log.Error(err, "unable to prepare the rest config for the target cluster", "cluster", cluster.Spec.Name)
		desc := fmt.Sprintf("unable to prepare the rest config for the target cluster due to error %s", err.Error())
		r.Recorder.Event(&cluster, v1.EventTypeWarning, string(state), desc)
		cluster.Status = managerv1alpha1.ClusterStatus{RetryCount: cluster.Status.RetryCount + 1, ErrorDescription: desc, State: state}
		return commonClient.UpdateStatus(ctx, &cluster, state, errRequeueTime)
	}

	// Isit being deleted?
	if cluster.ObjectMeta.DeletionTimestamp.IsZero() {
		//Good. This is not Delete use case
		//Lets check if this is very first time use case
		if !utils.ContainsString(cluster.ObjectMeta.Finalizers, clusterFinalizerName) {
			log.Info("New cluster resource. Adding the finalizer", "finalizer", clusterFinalizerName)
			cluster.ObjectMeta.Finalizers = append(cluster.ObjectMeta.Finalizers, clusterFinalizerName)
			commonClient.UpdateMeta(ctx, &cluster)
		}
		return r.HandleReconcile(ctx, req, &cluster, cfg)

	} else {
		//oh oh.. This is delete use case
		//Lets make sure to clean up the iam role
		if cluster.Status.RetryCount != 0 {
			cluster.Status.RetryCount = cluster.Status.RetryCount + 1
		}
		log.Info("Cluster delete request")
		if err := removeRBACInManagedCluster(ctx, cfg); err != nil {
			log.Error(err, "Unable to delete the cluster")
			cluster.Status = managerv1alpha1.ClusterStatus{RetryCount: cluster.Status.RetryCount + 1, ErrorDescription: err.Error(), State: managerv1alpha1.Error}
			commonClient.UpdateStatus(ctx, &cluster, managerv1alpha1.Error)
			r.Recorder.Event(&cluster, v1.EventTypeWarning, string(managerv1alpha1.Error), "unable to delete the cluster due to "+err.Error())
			return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
		}

		// Ok. Lets delete the finalizer so controller can delete the custom object
		log.Info("Removing finalizer from Cluster")
		cluster.ObjectMeta.Finalizers = utils.RemoveString(cluster.ObjectMeta.Finalizers, clusterFinalizerName)
		commonClient.UpdateMeta(ctx, &cluster)
		log.Info("Successfully deleted cluster")
		r.Recorder.Event(&cluster, v1.EventTypeNormal, "Deleted", "Successfully deleted cluster")
	}
	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) HandleReconcile(ctx context.Context, req ctrl.Request, cluster *managerv1alpha1.Cluster, cfg *rest.Config) (ctrl.Result, error) {
	log := log.Logger(ctx, "controllers", "cluster_controller", "HandleReconcile")
	log.WithValues("cluster_name", cluster.Spec.Name)
	log.Info("state of the custom resource ", "state", cluster.Status.State)
	commonClient := &common2.Client{Client: r.Client, Recorder: r.Recorder}

	state := managerv1alpha1.Warning

	if cluster.Status.RetryCount > 3 {
		state = managerv1alpha1.Error
	}
	// Validate the connection
	// Update the status
	// Requeue it based on config map variable
	// Ge the Managed cluster client
	resp, err := GetServerVersion(cfg)
	if err != nil {
		log.Error(err, "unable to get the server version", "cluster", cluster.Spec.Name)
		desc := fmt.Sprintf("Unable to get the server version due to error %s", err.Error())
		r.Recorder.Event(cluster, v1.EventTypeWarning, string(state), desc)
		cluster.Status = managerv1alpha1.ClusterStatus{RetryCount: cluster.Status.RetryCount + 1, ErrorDescription: desc, State: state}
		return commonClient.UpdateStatus(ctx, cluster, state, errRequeueTime)
	}
	r.Recorder.Event(cluster, v1.EventTypeNormal, string(managerv1alpha1.Ready), "Successfully validated the target cluster")
	cluster.Status = managerv1alpha1.ClusterStatus{RetryCount: 0, ErrorDescription: "", State: managerv1alpha1.Ready}
	commonClient.UpdateStatus(ctx, cluster, managerv1alpha1.Ready)
	log.Info("SUCCESSFUL", "version", resp)
	return ctrl.Result{RequeueAfter: time.Duration(config.Props.ClusterValidationFrequency()) * time.Second}, nil
}

func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&managerv1alpha1.Cluster{}).
		WithEventFilter(common2.StatusUpdatePredicate{}).
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

func removeRBACInManagedCluster(ctx context.Context, conf *rest.Config) error {
	log := log.Logger(ctx, "controllers", "cluster_controller", "removeRBACInManagedCluster")

	client := k8s.NewK8sManagedClusterClientDoOrDie(conf)

	// TO BE DISCUSSED: When you want to unregister any cluster, should we delete the service account as well??
	////Delete Cluster RoleBinding
	//err = client.DeleteClusterRoleBinding(ctx, common.ManagerClusterRoleBinding)
	//if err != nil {
	//	log.Error(err, "unable to delete cluster role binding in the target cluster")
	//	return err
	//}
	//
	////Delete Cluster Role
	//err = client.DeleteClusterRole(ctx, common.ManagerClusterRole)
	//if err != nil {
	//	log.Error(err, "unable to delete cluster role in the target cluster")
	//	return err
	//}

	//Delete Service Account
	err := client.DeleteServiceAccount(ctx, common.ManagerServiceAccountName, common.SystemNameSpace)
	if err != nil {
		log.Error(err, "unable to delete service account in the target cluster")
		return err
	}

	return nil
}
