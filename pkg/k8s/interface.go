package k8s

import (
	"context"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

//Interface defines required functions to be implemented by receivers
type Interface interface {
	SetUpEventHandler(ctx context.Context) record.EventRecorder
	GetConfigMap(ctx context.Context, ns string, name string) *v1.ConfigMap
	CreateServiceAccount(ctx context.Context, saName string, ns string) error
	DeleteServiceAccount(ctx context.Context, saName string, ns string) error
	CreateOrUpdateClusterRole(ctx context.Context, name string) error
	DeleteClusterRole(ctx context.Context, name string) error
	CreateOrUpdateClusterRoleBinding(ctx context.Context, name string) error
	DeleteClusterRoleBinding(ctx context.Context, name string) error
	GetServiceAccountTokenSecret(ctx context.Context, saName string) (string, error)
}
