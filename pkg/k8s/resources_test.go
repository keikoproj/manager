package k8s

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("K8s_rbac", func() {
	Describe("RBAC", func() {

		Context("New namespace creation", func() {
			It("should be successful", func() {
				Expect(cl.CreateNamespace(context.Background(), &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "valid-name"}})).To(BeNil())
			})
		})

		Context("New namespace creation with empty struct", func() {
			It("should fail as there is no name provided", func() {
				Expect(cl.CreateNamespace(context.Background(), &v1.Namespace{})).NotTo(BeNil())
			})
		})

		Context("New namespace creation with invalid DNS name", func() {
			It("should fail as there is name fails regex \"[a-z0-9]([-a-z0-9]*[a-z0-9])?\"", func() {
				Expect(cl.CreateNamespace(context.Background(), &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "valid.name"}})).NotTo(BeNil())
			})
		})

	})

})
