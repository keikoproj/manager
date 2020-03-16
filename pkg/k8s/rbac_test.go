package k8s

import (
	"context"
	"github.com/keikoproj/manager/internal/config/common"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/api/rbac/v1"
)

var _ = Describe("K8s_rbac", func() {
	Describe("RBAC", func() {

		Context("new service account creation", func() {
			It("should be successful", func() {
				Expect(cl.CreateServiceAccount(context.Background(), common.ManagerServiceAccountName, common.SystemNameSpace)).To(BeNil())
			})
		})
		Context("Trying to create the same service account", func() {
			It("shouldn't fail", func() {
				Expect(cl.CreateServiceAccount(context.Background(), common.ManagerServiceAccountName, common.SystemNameSpace)).To(BeNil())
			})
		})
		Context("new service account with bad name", func() {
			It("should throw error", func() {
				Expect(cl.CreateServiceAccount(context.Background(), "some_bad_name", common.SystemNameSpace)).NotTo(BeNil())
			})
		})

		Context("create new cluster role", func() {
			It("should be successful", func() {
				Expect(cl.CreateOrUpdateClusterRole(context.Background(), common.ManagerClusterRole)).To(BeNil())
			})
		})

		Context("Trying to create cluster role with same name", func() {
			It("should be successful", func() {
				Expect(cl.CreateOrUpdateClusterRole(context.Background(), common.ManagerClusterRole)).To(BeNil())
			})
		})

		Context("create new cluster role binding", func() {
			It("should be successful", func() {
				Expect(cl.CreateOrUpdateClusterRoleBinding(context.Background(), common.ManagerClusterRoleBinding, common.ManagerClusterRole, v1.Subject{
					Kind:      common.ServiceAccountKind,
					Name:      common.ManagerServiceAccountName,
					Namespace: common.SystemNameSpace,
				})).To(BeNil())
			})
		})

		Context("Trying to create cluster role binding with same name", func() {
			It("should be successful", func() {
				Expect(cl.CreateOrUpdateClusterRoleBinding(context.Background(), common.ManagerClusterRoleBinding, common.ManagerClusterRole, v1.Subject{
					Kind:      common.ServiceAccountKind,
					Name:      common.ManagerServiceAccountName,
					Namespace: common.SystemNameSpace,
				})).To(BeNil())
			})
		})

		Context("new cluster role binding with bad name", func() {
			It("should throw error", func() {
				Expect(cl.CreateOrUpdateClusterRoleBinding(context.Background(), "some_bad_name", "", v1.Subject{})).NotTo(BeNil())
			})
		})

		Context("Delete cluster role", func() {
			It("should be successful", func() {
				Expect(cl.DeleteClusterRole(context.Background(), common.ManagerClusterRole)).To(BeNil())
			})
		})

		Context("Delete cluster role which doesn't exist", func() {
			It("should be successful", func() {
				Expect(cl.DeleteClusterRole(context.Background(), "something-doesn't-exist")).To(BeNil())
			})
		})

		Context("Delete cluster role binding", func() {
			It("should be successful", func() {
				Expect(cl.DeleteClusterRoleBinding(context.Background(), common.ManagerClusterRoleBinding)).To(BeNil())
			})
		})

		Context("Delete cluster role binding which doesn't exists anymore", func() {
			It("should be successful", func() {
				Expect(cl.DeleteClusterRoleBinding(context.Background(), "something-doesn't-exist")).To(BeNil())
			})
		})

		Context("Delete cluster role binding which doesn't exists anymore", func() {
			It("should be successful", func() {
				Expect(cl.DeleteClusterRoleBinding(context.Background(), "something-doesn't-exist")).To(BeNil())
			})
		})

		Context("Delete service account", func() {
			It("should be successful", func() {
				Expect(cl.DeleteServiceAccount(context.Background(), common.ManagerServiceAccountName, common.SystemNameSpace)).To(BeNil())
			})
		})

		Context("Delete service account which doesn't exist anymore", func() {
			It("should be successful", func() {
				Expect(cl.DeleteServiceAccount(context.Background(), "doesn't-exist", common.SystemNameSpace)).To(BeNil())
			})
		})

		Context("Service token secret in testenv environment", func() {
			It("should throw error as testenv doesn't start with signing key", func() {
				_, err := cl.GetServiceAccountTokenSecret(context.Background(), common.ManagerServiceAccountName, common.SystemNameSpace)
				Expect(err).NotTo(BeNil())
			})
		})

	})

})
