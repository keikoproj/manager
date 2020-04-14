package k8s

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("resource.go testing", func() {
	Describe("Namespace creation", func() {

		Context("New namespace creation", func() {
			It("should be successful", func() {
				Expect(cl.CreateOrUpdateNamespace(context.Background(), &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "valid-name"}})).To(BeNil())
			})
		})

		Context("New namespace creation with empty struct", func() {
			It("should fail as there is no name provided", func() {
				Expect(cl.CreateOrUpdateNamespace(context.Background(), &v1.Namespace{})).NotTo(BeNil())
			})
		})

		Context("New namespace creation with invalid DNS name", func() {
			It("should fail as there is name fails regex \"[a-z0-9]([-a-z0-9]*[a-z0-9])?\"", func() {
				Expect(cl.CreateOrUpdateNamespace(context.Background(), &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "valid.name"}})).NotTo(BeNil())
			})
		})

	})

	Describe("Resource Quota creation", func() {

		Context("New resource quota creation valid use case", func() {
			It("should be successful", func() {
				Expect(cl.CreateOrUpdateResourceQuota(context.Background(), &v1.ResourceQuota{
					ObjectMeta: metav1.ObjectMeta{Name: "valid-resource-quota"},
					Spec: v1.ResourceQuotaSpec{
						Hard: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("3"),
							v1.ResourceMemory: resource.MustParse("100Gi"),
							v1.ResourcePods:   resource.MustParse("5"),
						},
					}}, "valid-name")).To(BeNil())
			})
		})

		Context("New resource quota creation invalid name", func() {
			It("should be successful", func() {
				Expect(cl.CreateOrUpdateResourceQuota(context.Background(), &v1.ResourceQuota{
					ObjectMeta: metav1.ObjectMeta{Name: "valid-resource-Quota"},
					Spec: v1.ResourceQuotaSpec{
						Hard: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("3"),
							v1.ResourceMemory: resource.MustParse("100Gi"),
							v1.ResourcePods:   resource.MustParse("5"),
						},
					}}, "valid-name")).ToNot(BeNil())
			})
		})

	})

})
