package k8s

import (
	"context"
	"github.com/keikoproj/manager/pkg/grpc/proto/namespace"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("custom.go testing", func() {
	Describe("Custom Resource creation", func() {
		//define namespace cr
		cr := &namespace.CustomResource{
			GVK: &namespace.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Namespace",
			},
		}
		jsonString := `{
  "apiVersion": "v1",
  "kind": "Namespace",
  "metadata": {
    "name": "docker-desktop"
  }
}`
		cr.Manifest = jsonString
		Context("New namespace creation using controller_runtime client", func() {
			It("should be successful", func() {
				Expect(cl.CreateOrUpdateCustomResource(context.Background(), cr, "docker-desktop")).To(BeNil())
			})
		})

		Context("namespace update using controller_runtime client", func() {
			It("should be successful", func() {
				Expect(cl.CreateOrUpdateCustomResource(context.Background(), cr, "docker-desktop")).To(BeNil())
			})
		})
		Context("Invalid json as a manifest string", func() {
			It("should throw err", func() {
				cr.Manifest = "something"
				Expect(cl.CreateOrUpdateCustomResource(context.Background(), cr, "docker_desktop")).ToNot(BeNil())
			})
		})
		Context("Invalid manifest as a string", func() {
			It("should throw err", func() {
				cr.Manifest = `{
  "apiVersion": "v1",
  "kind": "Namespace",
  "metadata": {
    "name": "docker.desktop"
  }
}`
				Expect(cl.CreateOrUpdateCustomResource(context.Background(), cr, "docker_desktop")).ToNot(BeNil())
			})
		})
	})

})
