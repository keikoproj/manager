package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/keikoproj/manager/internal/utils"
	"github.com/keikoproj/manager/pkg/k8s"
	"github.com/keikoproj/manager/pkg/log"
	"github.com/pborman/uuid"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	managerv1alpha1 "github.com/keikoproj/manager/api/v1alpha1"
	common2 "github.com/keikoproj/manager/controllers/common"
	controllercommon "github.com/keikoproj/manager/controllers/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	applicationFinalizerName = "application.finalizers.manager.keikoproj.io"
)

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Log           logr.Logger
	Scheme        *runtime.Scheme
	K8sSelfClient *k8s.Client
	Recorder      record.EventRecorder
}

// +kubebuilder:rbac:groups=manager.keikoproj.io,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=manager.keikoproj.io,resources=applications/status,verbs=get;update;patch
func (r *ApplicationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	ctx := context.WithValue(context.Background(), requestId, uuid.New())
	log := log.Logger(ctx, "controllers", "application_controller", "Reconcile")
	log = log.WithValues("namespace", req.NamespacedName)
	log.Info("Start of the request")
	//Get the resource
	var app managerv1alpha1.Application

	if err := r.Get(ctx, req.NamespacedName, &app); err != nil {
		return ctrl.Result{}, ignoreNotFound(err)
	}
	commonClient := &controllercommon.Client{Client: r.Client, Recorder: r.Recorder, K8sSelfClient: r.K8sSelfClient}

	// Isit being deleted?
	if app.ObjectMeta.DeletionTimestamp.IsZero() {
		//Good. This is not Delete use case
		if !utils.ContainsString(app.ObjectMeta.Finalizers, applicationFinalizerName) {
			log.Info("New application resource. Adding the finalizer", "finalizer", applicationFinalizerName)
			app.ObjectMeta.Finalizers = append(app.ObjectMeta.Finalizers, applicationFinalizerName)
			commonClient.UpdateMeta(ctx, &app)
		}

		return r.HandleReconcile(ctx, &app)

	} else {
		//oh oh.. This is delete use case
		//Lets make sure to clean up the iam role
		if app.Status.RetryCount != 0 {
			app.Status.RetryCount = app.Status.RetryCount + 1
		}
		log.Info("Application delete request")

		// Ok. Lets delete the finalizer so controller can delete the custom object
		log.Info("Removing finalizer from application")
		app.ObjectMeta.Finalizers = utils.RemoveString(app.ObjectMeta.Finalizers, applicationFinalizerName)
		commonClient.UpdateMeta(ctx, &app)
		log.Info("Successfully deleted application")
		r.Recorder.Event(&app, v1.EventTypeNormal, "Deleted", "Successfully deleted application")
	}

	return ctrl.Result{}, nil
}

func (r *ApplicationReconciler) HandleReconcile(ctx context.Context, app *managerv1alpha1.Application) (ctrl.Result, error) {
	log := log.Logger(ctx, "controllers", "application_controller", "HandleReconcile")
	commonClient := &controllercommon.Client{Client: r.Client, Recorder: r.Recorder, K8sSelfClient: r.K8sSelfClient}

	// lets see what we should do
	// we should just create the mns from here
	// only thing i gotta do is, add the params to mns params

	//This will handle the (common) params to be propagated to namespaces
	for k, v := range app.Spec.AppParams {
		for _, env := range app.Spec.Environments {
			if _, ok := env.Namespace.Params[k]; !ok {
				env.Namespace.Params[k] = v
			}
		}
	}

	for _, env := range app.Spec.Environments {
		//Lets construct the name based on appName and env
		//TODO: Allow overwriting the namespace name in future
		name := fmt.Sprintf("%s-%s", app.Spec.AppName, env.Name)
		log = log.WithValues("namespaceName", name)
		mns := &managerv1alpha1.ManagedNamespace{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: utils.SanitizeName(env.Namespace.ClusterName),
			},
			Spec: managerv1alpha1.ManagedNamespaceSpec{
				Namespace: *env.Namespace,
			},
		}

		if err := ctrl.SetControllerReference(app, mns, r.Scheme); err != nil {
			log.Error(err, "Unable to set the controller reference")
			desc := fmt.Sprintf("Unable to set the controller reference due to error %s", err.Error())
			r.Recorder.Event(app, v1.EventTypeWarning, string(managerv1alpha1.Error), desc)
			app.Status = managerv1alpha1.ApplicationStatus{RetryCount: app.Status.RetryCount + 1, ErrorDescription: desc, State: managerv1alpha1.Error}
			return commonClient.UpdateStatus(ctx, app, managerv1alpha1.Error, errRequeueTime)
		}

		//Apply patch
		//applyOpts := []client.PatchOption{client.ForceOwnership, client.FieldOwner("application-controller")}

		err := r.K8sSelfClient.CreateOrUpdateManagedNamespace(ctx, mns, utils.SanitizeName(env.Namespace.ClusterName))
		if err != nil {
			log.Error(err, "Unable to create managed namespace")
			desc := fmt.Sprintf("Unable to create managed namespace due to error %s", err.Error())
			r.Recorder.Event(app, v1.EventTypeWarning, string(managerv1alpha1.Error), desc)
			app.Status = managerv1alpha1.ApplicationStatus{RetryCount: app.Status.RetryCount + 1, ErrorDescription: desc, State: managerv1alpha1.Error}
			return commonClient.UpdateStatus(ctx, app, managerv1alpha1.Error, errRequeueTime)
		}

		log.Info("Successfully created managed namespace")

	}

	log.Info("Successfully created application", "appName", app.Spec.AppName)
	r.Recorder.Event(app, v1.EventTypeNormal, string(managerv1alpha1.Ready), "Successfully created/updated application")
	app.Status = managerv1alpha1.ApplicationStatus{RetryCount: 0, ErrorDescription: "", State: managerv1alpha1.Ready}
	commonClient.UpdateStatus(ctx, app, managerv1alpha1.Ready)

	return ctrl.Result{}, nil
}

func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&managerv1alpha1.Application{}).
		WithEventFilter(common2.StatusUpdatePredicate{}).
		Complete(r)
}
