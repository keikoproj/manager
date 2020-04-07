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
	"github.com/go-logr/logr"
	managerv1alpha1 "github.com/keikoproj/manager/api/custom/v1alpha1"
	controllercommon "github.com/keikoproj/manager/controllers/common"
	"github.com/keikoproj/manager/internal/config/common"
	"github.com/keikoproj/manager/internal/utils"
	"github.com/keikoproj/manager/pkg/grpc/proto/namespace"
	"github.com/keikoproj/manager/pkg/k8s"
	"github.com/keikoproj/manager/pkg/log"
	"github.com/keikoproj/manager/pkg/template"
	"github.com/pborman/uuid"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	namespaceFinalizerName = "namespace.finalizers.manager.keikoproj.io"
)

// ManagedNamespaceReconciler reconciles a ManagedNamespace object
type ManagedNamespaceReconciler struct {
	client.Client
	Log           logr.Logger
	Scheme        *runtime.Scheme
	Recorder      record.EventRecorder
	K8sSelfClient *k8s.Client
}

// +kubebuilder:rbac:groups=manager.keikoproj.io,resources=managednamespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=manager.keikoproj.io,resources=managednamespaces/status,verbs=get;update;patch

func (r *ManagedNamespaceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.WithValue(context.Background(), requestId, uuid.New())
	log := log.Logger(ctx, "controllers", "namespace_controller", "Reconcile")
	log = log.WithValues("namespace", req.NamespacedName)
	log.Info("Start of the request")
	//Get the resource
	var ns managerv1alpha1.ManagedNamespace

	if err := r.Get(ctx, req.NamespacedName, &ns); err != nil {
		return ctrl.Result{}, ignoreNotFound(err)
	}
	commonClient := &controllercommon.Client{Client: r.Client, Recorder: r.Recorder, K8sSelfClient: r.K8sSelfClient}

	// Final template to be processed
	if err := r.FinalNSTemplate(ctx, &ns); err != nil {
		log.Error(err, "unable to process namespace template", "template", ns.Spec.TemplateName)
		desc := fmt.Sprintf("unable to process namespace template due to error %s", err.Error())
		r.Recorder.Event(&ns, v1.EventTypeWarning, string(managerv1alpha1.Error), desc)
		ns.Status = managerv1alpha1.ManagedNamespaceStatus{RetryCount: ns.Status.RetryCount + 1, ErrorDescription: desc, State: managerv1alpha1.Error}
		return commonClient.UpdateStatus(ctx, &ns, managerv1alpha1.Error, errRequeueTime)
	}

	//K8s client for the managed cluster
	k8sManagedClient, err := r.ManagedClusterClient(ctx, &ns)
	if err != nil {
		log.Error(err, "unable to get the cluster details for the namespace")
		desc := fmt.Sprintf("unable to get the cluster details for the namespace due to error %s", err.Error())
		r.Recorder.Event(&ns, v1.EventTypeWarning, string(managerv1alpha1.Error), desc)
		ns.Status = managerv1alpha1.ManagedNamespaceStatus{RetryCount: ns.Status.RetryCount + 1, ErrorDescription: desc, State: managerv1alpha1.Error}
		return commonClient.UpdateStatus(ctx, &ns, managerv1alpha1.Error, errRequeueTime)
	}

	// Isit being deleted?
	if ns.ObjectMeta.DeletionTimestamp.IsZero() {
		//Good. This is not Delete use case
		//Lets check if this is very first time use case
		firstTime := false
		if !utils.ContainsString(ns.ObjectMeta.Finalizers, namespaceFinalizerName) {
			log.Info("New managed namespace resource. Adding the finalizer", "finalizer", namespaceFinalizerName)
			ns.ObjectMeta.Finalizers = append(ns.ObjectMeta.Finalizers, namespaceFinalizerName)
			firstTime = true
			commonClient.UpdateMeta(ctx, &ns)
		}
		return r.HandleNSResources(ctx, &ns, k8sManagedClient, firstTime)

	} else {
		//oh oh.. This is delete use case
		//Lets make sure to clean up the iam role
		if ns.Status.RetryCount != 0 {
			ns.Status.RetryCount = ns.Status.RetryCount + 1
		}
		log.Info("Namespace delete request")

		// Ok. Lets delete the finalizer so controller can delete the custom object
		log.Info("Removing finalizer from managed namespace")
		ns.ObjectMeta.Finalizers = utils.RemoveString(ns.ObjectMeta.Finalizers, namespaceFinalizerName)
		commonClient.UpdateMeta(ctx, &ns)
		log.Info("Successfully deleted managed namespace")
		r.Recorder.Event(&ns, v1.EventTypeNormal, "Deleted", "Successfully deleted managed namespace")
	}

	//log.Info("Successfully reconciled managed namespace resource", "name", ns.Spec.Name)
	return ctrl.Result{}, nil
}

//ResourceStatus represents each resource status
type ResourceStatus struct {
	Name      string
	Type      string
	DependsOn string
	Done      bool
	Error     error
}

//HandleNSResources manages namespaces resources
func (r *ManagedNamespaceReconciler) HandleNSResources(ctx context.Context, ns *managerv1alpha1.ManagedNamespace, k8sManagedClient *k8s.Client, firstTime bool) (ctrl.Result, error) {
	log := log.Logger(ctx, "controllers", "managednamespace_controller", "HandleNSResources")

	commonClient := &controllercommon.Client{Client: r.Client, Recorder: r.Recorder}

	err := k8sManagedClient.CreateOrUpdateNamespace(ctx, ns.Spec.NsResources.Namespace)
	if err != nil {
		log.Error(err, "unable to create the namespace", "cluster", ns.Spec.ClusterName, "ns", ns.Spec.ClusterName)
		desc := fmt.Sprintf("unable to create the namespace due to error %s", err.Error())
		r.Recorder.Event(ns, v1.EventTypeWarning, string(managerv1alpha1.Error), desc)
		ns.Status = managerv1alpha1.ManagedNamespaceStatus{RetryCount: ns.Status.RetryCount + 1, ErrorDescription: desc, State: managerv1alpha1.Error}
		return commonClient.UpdateStatus(ctx, ns, managerv1alpha1.Error, errRequeueTime)
	}
	statusMap := make(map[string]ResourceStatus)
	var ResourceFunction func() (int, error)
	x := 0
	ResourceFunction = func() (int, error) {
		if len(ns.Spec.NsResources.Resources) == x {
			return 0, nil
		}
		n := 0
		Error := make(chan error)
		Done := make(chan bool)
		for _, res := range ns.Spec.NsResources.Resources {

			if shouldProceed(ctx, statusMap, res, firstTime) {
				n++
				go func(statusMap map[string]ResourceStatus, res *namespace.Resource) {
					//This is where i have to do switch case check
					var err error
					// Check whether
					switch res.Type {
					case common.ServiceAccountKind:
						log.V(1).Info("Service Account creation is in progress", "name", res.ServiceAccount.Name)
						err = k8sManagedClient.CreateServiceAccount(ctx, res.ServiceAccount, ns.Spec.NsResources.Namespace.Name)

					case common.RoleKind:
						log.V(1).Info("Role creation is in progress")
						err = k8sManagedClient.CreateOrUpdateRole(ctx, res.Role, ns.Spec.NsResources.Namespace.Name)

					case common.RoleBindingKind:
						log.V(1).Info("RoleBinding creation is in progress", "name", res.RoleBinding.Name)
						err = k8sManagedClient.CreateOrUpdateRoleBinding(ctx, res.RoleBinding, ns.Spec.NsResources.Namespace.Name)

					case common.ResourceQuotaKind:
						log.V(1).Info("Resource Quota creation is in progress")
						err = k8sManagedClient.CreateOrUpdateResourceQuota(ctx, res.ResourceQuota, ns.Spec.NsResources.Namespace.Name)

					default:
						//TODO: handle error management
						log.Info("Invalid choice")

					}
					status := statusMap[res.Name]
					if err != nil {
						log.Error(err, "unable to create the resource", "name", res.Name, "type", res.Type)
						desc := fmt.Sprintf("unable to create the resource due to error %s", err.Error())
						r.Recorder.Event(ns, v1.EventTypeWarning, string(managerv1alpha1.Error), desc)
						//ns.Status = managerv1alpha1.ManagedNamespaceStatus{RetryCount: ns.Status.RetryCount + 1, ErrorDescription: desc, State: managerv1alpha1.Error}
						//commonClient.UpdateStatus(ctx, ns, managerv1alpha1.Error, errRequeueTime)
						status.Error = err
						statusMap[res.Name] = status
						Error <- err

					} else {
						status.Done = true
						statusMap[res.Name] = status
						desc := fmt.Sprintf("successfully created/updated resource %s", res.Name)
						r.Recorder.Event(ns, v1.EventTypeNormal, string(managerv1alpha1.Ready), desc)
						Done <- true
					}

				}(statusMap, res)
			}

		}
		i := 0
		for {
			select {
			case err = <-Error:
				i++
			case <-Done:
				i++
			}
			//TODO: Figure out a way to handle
			if i == n {
				break
			}
		}
		// Close channel
		close(Error)
		close(Done)

		x = x + n
		if err != nil {
			//Lets not proceed further and do the recursive function
			log.Error(err, "unable to create one of the resource. aborting")
			return x, err
		}
		//recursive call if all the resources are not done
		return ResourceFunction()
	}

	count, err := ResourceFunction()
	log.Info("Total Resources created", "count", count)
	if err != nil {
		log.Error(err, "unable to create one of the resource. aborting")
		desc := fmt.Sprintf("unable to create the resource due to error %s", err.Error())
		r.Recorder.Event(ns, v1.EventTypeWarning, string(managerv1alpha1.Error), desc)
		ns.Status = managerv1alpha1.ManagedNamespaceStatus{RetryCount: ns.Status.RetryCount + 1, ErrorDescription: desc, State: managerv1alpha1.Error}
		return commonClient.UpdateStatus(ctx, ns, managerv1alpha1.Error, errRequeueTime)
	}

	return ctrl.Result{}, nil
}

func (r *ManagedNamespaceReconciler) FinalNSTemplate(ctx context.Context, ns *managerv1alpha1.ManagedNamespace) error {
	// Figure out the default template (Should i update the resource?)
	log := log.Logger(ctx, "controllers", "namespace_controller", "Reconcile")
	//
	// If template name not included in the request
	if ns.Spec.TemplateName == "" {
		log.V(1).Info("No template name included")
		return nil
	}
	//Lets see if we can get the namespace template
	log.V(1).Info("Retrieving namespace template")
	var nsTemplate managerv1alpha1.NamespaceTemplate
	templateNamespacedName := types.NamespacedName{Namespace: "", Name: ns.Spec.TemplateName}
	if err := r.Get(ctx, templateNamespacedName, &nsTemplate); err != nil {
		log.Error(err, "unable to get the namespace template requested", "template", ns.Spec.TemplateName)
		return err
	}

	if err := template.ProcessTemplate(ctx, &nsTemplate, ns); err != nil {
		log.Error(err, "unable to process namespace template", "template", ns.Spec.TemplateName)
		return err
	}
	return nil
}

func (r *ManagedNamespaceReconciler) ManagedClusterClient(ctx context.Context, ns *managerv1alpha1.ManagedNamespace) (*k8s.Client, error) {
	log := log.Logger(ctx, "controllers", "namespace_controller", "Reconcile")
	commonClient := &controllercommon.Client{Client: r.Client, Recorder: r.Recorder, K8sSelfClient: r.K8sSelfClient}
	// Get the cluster
	log.V(1).Info("Retrieving cluster info")
	var cluster managerv1alpha1.Cluster
	clusterNamespacedName := types.NamespacedName{Namespace: ns.ObjectMeta.Namespace, Name: ns.Spec.ClusterName}
	if err := r.Get(ctx, clusterNamespacedName, &cluster); err != nil {
		log.Error(err, "unable to get the cluster details for the namespace", "cluster", cluster.Spec.Name)
		return nil, err
	}

	log.V(1).Info("Cluster info", "secretName", cluster.Spec.Config.BearerTokenSecret)

	k8sManagedClient, err := commonClient.ManagedClusterK8sClient(ctx, &cluster)
	if err != nil {
		log.Error(err, "unable to get the cluster details for the namespace", "cluster", cluster.Spec.Name)
		return nil, err
	}
	return k8sManagedClient, nil
}

func (r *ManagedNamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&managerv1alpha1.ManagedNamespace{}).
		WithEventFilter(controllercommon.StatusUpdatePredicate{}).
		Complete(r)
}

//shouldProceed checks whether to proceed further
func shouldProceed(ctx context.Context, statusMap map[string]ResourceStatus, resource *namespace.Resource, firstTime bool) bool {
	log := log.Logger(ctx, "controllers", "managednamespace_controller", "shouldProceed")
	log = log.WithValues("resource", resource.Name)
	proceed := false

	if utils.BoolValue(resource.CreateOnly) && !firstTime {
		proceed = false
		log.V(1).Info("result", "proceed", proceed, "firstTime", firstTime)
		return proceed
	}

	if val, ok := statusMap[resource.Name]; ok {

		if val.Done {
			proceed = false
			log.V(1).Info("result", "proceed", proceed)
			return proceed
		}

		// This means its not the first iteration but it has DependsOn value
		if val.DependsOn != "" {
			proceed = false
			if statusMap[val.DependsOn].Done {
				proceed = true
				log.V(1).Info("result", "proceed", proceed)
				return proceed
			}
			log.V(1).Info("result", "proceed", proceed)
			return proceed
		}
		if val.Error != nil {
			proceed = false
			log.V(1).Info("THIS SHOULDN't BE THE CASE", "proceed", proceed)

			//This shouldn't be the case and SHOULD PROBABLY FREAK OUT
			return proceed
		}

	}

	//This means first iteration
	// Lets create the status and add it to the map
	statusMap[resource.Name] = ResourceStatus{
		Name:      resource.Name,
		Type:      resource.Type,
		DependsOn: resource.DependsOn,
	}

	//Check whether dependsOn has any value
	if resource.DependsOn == "" {
		proceed = true
		//This means doesn't dependOn anything
		// it can proceed
		log.V(1).Info("First iteration", "proceed", proceed)

		return proceed
	}

	return false
}
