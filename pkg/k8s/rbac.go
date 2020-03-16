package k8s

import (
	"context"
	"errors"
	"fmt"
	"github.com/keikoproj/manager/internal/config/common"
	"github.com/keikoproj/manager/pkg/log"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//CreateServiceAccount adds the service account in the target cluster
func (c *Client) CreateServiceAccount(ctx context.Context, saName string, ns string) error {
	log := log.Logger(ctx, "pkg.k8s", "rbac", "CreateServiceAccount")
	serviceAccount := corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       common.ServiceAccountKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      saName,
			Namespace: ns,
		},
	}
	_, err := c.cl.CoreV1().ServiceAccounts(ns).Create(&serviceAccount)
	if err != nil {
		if !apierr.IsAlreadyExists(err) {
			msg := fmt.Sprintf("Failed to create service account %s in namespace %s due to %v", saName, ns, err)
			log.Error(err, msg)
			return errors.New(msg)
		}
		log.Info("Service account already exists", "serviceAccount", saName, "namespace", ns)
		return nil
	}
	log.Info("Service account got created successfully", "serviceAccount", saName, "namespace", ns)
	return nil
}

//DeleteServiceAccount deletes the service account in the target cluster
func (c *Client) DeleteServiceAccount(ctx context.Context, saName string, ns string) error {
	log := log.Logger(ctx, "pkg.k8s", "client", "DeleteServiceAccount")

	err := c.cl.CoreV1().ServiceAccounts(ns).Delete(saName, &metav1.DeleteOptions{})
	if err != nil {
		if !apierr.IsNotFound(err) {
			msg := fmt.Sprintf("Failed to delete service account %s in namespace %s due to %v", saName, ns, err)
			log.Error(err, msg)
			return errors.New(msg)
		}
		log.Info("Service account doesn't exists anymore", "serviceAccount", saName, "namespace", ns)
		return nil
	}
	log.Info("Service account removed successfully", "serviceAccount", saName, "namespace", ns)
	return nil
}

//CreateOrUpdateClusterRole create or updates cluster role
func (c *Client) CreateOrUpdateClusterRole(ctx context.Context, name string) error {
	log := log.Logger(ctx, "pkg.k8s", "client", "AddServiceAccount")
	clusterRole := rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: common.RBACApiVersion,
			Kind:       common.ClusterRoleKind,
		},

		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
			{
				NonResourceURLs: []string{"*"},
				Verbs:           []string{"*"},
			},
		},
	}

	_, err := c.cl.RbacV1().ClusterRoles().Create(&clusterRole)
	if err != nil {
		if !apierr.IsAlreadyExists(err) {
			msg := fmt.Sprintf("Failed to create cluster role %s due to %v", name, err)
			log.Error(err, msg)
			return errors.New(msg)
		}
		log.Info("Cluster Role Already exists. Trying to update", "clusterRole", name)
		//Already exists. lets Update it
		_, err := c.cl.RbacV1().ClusterRoles().Update(&clusterRole)
		if err != nil {
			msg := fmt.Sprintf("Failed to update cluster role %s due to %v", name, err)
			log.Error(err, msg)
			return errors.New(msg)
		}
	}
	log.Info("Successfully created cluster role", "clusterRole", name)
	return nil
}

//DeleteClusterRole deletes cluster role
func (c *Client) DeleteClusterRole(ctx context.Context, name string) error {
	log := log.Logger(ctx, "pkg.k8s", "client", "DeleteClusterRole")

	err := c.cl.RbacV1().ClusterRoles().Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		if !apierr.IsNotFound(err) {
			msg := fmt.Sprintf("Failed to delete cluster role %s due to %v", name, err)
			log.Error(err, msg)
			return errors.New(msg)
		}
		log.Info("Cluster Role doesn't exist anymore", "clusterRole", name)
	}
	log.Info("Successfully removed cluster role", "clusterRole", name)
	return nil
}

//CreateOrUpdateClusterRole create or updates cluster role
func (c *Client) CreateOrUpdateClusterRoleBinding(ctx context.Context, name string, clusterRoleName string, subject rbacv1.Subject) error {
	log := log.Logger(ctx, "pkg.k8s", "client", "CreateOrUpdateClusterRoleBinding")
	clusterRoleBinding := rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: common.RBACApiVersion,
			Kind:       common.ClusterRoleBindingKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     common.ClusterRoleKind,
			Name:     clusterRoleName,
		},
		Subjects: []rbacv1.Subject{subject},
	}

	_, err := c.cl.RbacV1().ClusterRoleBindings().Create(&clusterRoleBinding)
	if err != nil {
		if !apierr.IsAlreadyExists(err) {
			msg := fmt.Sprintf("Failed to create cluster role binding %s due to %v", name, err)
			log.Error(err, msg)
			return errors.New(msg)
		}
		log.Info("Cluster RoleBinding Already exists. Trying to update", "clusterRoleBinding", name, "clusterRole", clusterRoleName)
		//Already exists. lets Update it
		_, err := c.cl.RbacV1().ClusterRoleBindings().Update(&clusterRoleBinding)
		if err != nil {
			msg := fmt.Sprintf("Failed to update cluster role binding %s due to %v", name, err)
			log.Error(err, msg)
			return errors.New(msg)
		}
	}
	log.Info("Successfully created cluster RoleBinding", "clusterRoleBinding", name, "clusterRole", clusterRoleName)
	return nil
}

//CreateOrUpdateClusterRole create or updates cluster role
func (c *Client) DeleteClusterRoleBinding(ctx context.Context, name string) error {
	log := log.Logger(ctx, "pkg.k8s.rbac", "DeleteClusterRoleBinding")

	err := c.cl.RbacV1().ClusterRoleBindings().Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		if !apierr.IsNotFound(err) {
			msg := fmt.Sprintf("Failed to delete cluster role binding %s due to %v", name, err)
			log.Error(err, msg)
			return errors.New(msg)
		}
		log.Info("Cluster RoleBinding doesn't exist anymore", "clusterRoleBinding", name)
	}
	log.Info("Successfully removed Cluster RoleBinding", "clusterRoleBinding", name)
	return nil
}

//GetServiceAccountTokenSecret retrieves the token secret for a given service account
func (c *Client) GetServiceAccountTokenSecret(ctx context.Context, saName string, ns string) (string, error) {
	log := log.Logger(ctx, "pkg.k8s.rbac", "GetServiceAccountTokenSecret")

	var secret *corev1.Secret
	err := wait.Poll(500*time.Millisecond, 300*time.Second, func() (bool, error) {
		sa, err := c.cl.CoreV1().ServiceAccounts(ns).Get(saName, metav1.GetOptions{})
		if err != nil {
			log.Error(err, "unable to retrieve service account", "serviceAccountName", saName)
			return false, err
		}

		for _, obj := range sa.Secrets {
			secret, err = c.cl.CoreV1().Secrets(ns).Get(obj.Name, metav1.GetOptions{})
			log.Info("secret found", "secret_name", secret.Name)
			if err != nil {
				log.Error(err, "unable to retrieve service account secret", "serviceAccountName", saName, "secretName", obj.Name)
				return false, err
			}
			if secret.Type == corev1.SecretTypeServiceAccountToken {
				return true, nil
			}
		}
		//it shouldn't reach until here
		return false, nil
	})

	if err != nil {
		log.Error(err, "unable to retrieve service account secret", "serviceAccountName", saName)
		return "", err
	}

	token, ok := secret.Data["token"]
	if !ok {
		msg := "service account secret doesn't have token"
		log.Error(err, msg, "serviceAccountName", saName)

		return "", errors.New(msg)
	}
	log.V(1).Info("service account token secret retrieved successfully")
	return string(token), nil
}
