package k8s

import (
	"context"
	"github.com/keikoproj/manager/api/custom/v1alpha1"
	"github.com/keikoproj/manager/pkg/grpc/proto/namespace"
	"k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/client-go/tools/record"
)

//Interface defines required functions to be implemented by receivers
type Interface interface {
	SetUpEventHandler(ctx context.Context) record.EventRecorder
	GetConfigMap(ctx context.Context, ns string, name string) *v1.ConfigMap
	CreateServiceAccountForCluster(ctx context.Context, saName string, ns string) error
	CreateServiceAccount(ctx context.Context, sa *v1.ServiceAccount) error
	DeleteServiceAccount(ctx context.Context, saName string, ns string) error
	CreateOrUpdateClusterRole(ctx context.Context, name string) error
	DeleteClusterRole(ctx context.Context, name string) error
	CreateOrUpdateClusterRoleBinding(ctx context.Context, name string) error
	DeleteClusterRoleBinding(ctx context.Context, name string) error

	CreateOrUpdateRole(ctx context.Context, role *rbacv1.Role, ns string) error
	CreateOrUpdateRoleBinding(ctx context.Context, roleBinding *rbacv1.RoleBinding) error

	GetServiceAccountTokenSecret(ctx context.Context, saName string) (string, error)
	CreateOrUpdateK8sSecret(ctx context.Context, secret *v1.Secret) error
	GetK8sSecret(ctx context.Context, name string, ns string) (*v1.Secret, error)

	CreateOrUpdateNamespace(ctx context.Context, namespace *v1.Namespace) error
	DeleteNamespace(ctx context.Context, name string) error
	GetNamespace(ctx context.Context, name string) error

	CreateOrUpdateResourceQuota(ctx context.Context, quota *v1.ResourceQuota) error

	CreateOrUpdateClusterCR(ctx context.Context, cr *v1alpha1.Cluster) error
	CreateOrUpdateCustomResource(ctx context.Context, cr *namespace.CustomResource, ns string) error
	CreateOrUpdateManagedNamespace(ctx context.Context, cr *v1alpha1.ManagedNamespace, ns string) error
}
