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
	"github.com/keikoproj/manager/pkg/log"
	"github.com/pborman/uuid"

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
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=manager.keikoproj.io,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=manager.keikoproj.io,resources=clusters/status,verbs=get;update;patch

func (r *ClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	ctx := context.WithValue(context.Background(), requestId, uuid.New())
	log := log.Logger(ctx, "controllers", "iamrole_controller", "Reconcile")
	log.WithValues("iamrole", req.NamespacedName)
	log.Info("Start of the request")
	//Get the resource
	var cluster managerv1alpha1.Cluster
	if err := r.Get(ctx, req.NamespacedName, &cluster); err != nil {
		return ctrl.Result{}, ignoreNotFound(err)
	}

	log.Info("received", "cluster_secret", cluster.Spec.Config.BearerTokenSecret)

	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&managerv1alpha1.Cluster{}).
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
