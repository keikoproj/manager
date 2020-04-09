package template_test

import (
	"context"
	"github.com/keikoproj/manager/api/custom/v1alpha1"
	"github.com/keikoproj/manager/pkg/grpc/proto/namespace"
	"github.com/keikoproj/manager/pkg/template"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("template test suite", func() {
	//TODO: Add more test cases with custom resources
	Describe("ProcessTemplate test cases", func() {
		//manifest := `{\"apiVersion\":\"v1\",\"kind\":\"ServiceAccount\",\"metadata\":{\"name\":\"preprod-sa3\",\"namespace\":\"second-namespace\"}}`
		nsTemplate := namespace.NamespaceTemplate{
			ExportedParamName: []string{"env"},
			NsResources: &namespace.NamespaceResources{
				Namespace: &v1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: "local_namespace",
					},
				},
				Resources: []*namespace.Resource{
					{
						Type: "ServiceAccount",
						Name: "local_sa",
						ServiceAccount: &v1.ServiceAccount{
							ObjectMeta: metav1.ObjectMeta{
								Name: "local_sa",
							},
						},
					},
				},
			},
		}
		params := make(map[string]string)
		params["env"] = "preprod"
		nsReq := namespace.Namespace{
			Params: params,
		}
		mns := &v1alpha1.ManagedNamespace{
			Spec: v1alpha1.ManagedNamespaceSpec{
				Namespace: nsReq,
			},
		}
		Context("Number of Resources usecase", func() {
			It("error should be nil", func() {
				Expect(template.ProcessTemplate(context.Background(), &v1alpha1.NamespaceTemplate{
					Spec: v1alpha1.NamespaceTemplateSpec{
						NamespaceTemplate: nsTemplate,
					},
				}, mns)).To(BeNil())
				Expect(len(mns.Spec.NsResources.Resources)).To(Equal(1))
			})

		})

	})
	Describe("custom resource template test cases", func() {
		manifest := `{\"apiVersion\":\"v1\",\"kind\":\"ServiceAccount\",\"metadata\":{\"name\":\"preprod-sa3\",\"namespace\":\"second-namespace\"}}`
		nsTemplate := namespace.NamespaceTemplate{
			ExportedParamName: []string{"env"},
			NsResources: &namespace.NamespaceResources{
				Resources: []*namespace.Resource{
					{
						CustomResource: &namespace.CustomResource{
							Manifest: manifest,
						},
						Type: "CustomResource",
						Name: "something",
					},
				},
			},
		}
		params := make(map[string]string)
		params["env"] = "preprod"
		nsReq := namespace.Namespace{
			Params: params,
			NsResources: &namespace.NamespaceResources{
				Resources: []*namespace.Resource{},
			},
		}
		Context("manifest in json string quote escaped from input", func() {
			It("should be able to handle", func() {
				Expect(template.ProcessCustomResourceTemplate(context.Background(), v1alpha1.NamespaceTemplate{
					Spec: v1alpha1.NamespaceTemplateSpec{
						NamespaceTemplate: nsTemplate,
					},
				}, &v1alpha1.ManagedNamespace{
					Spec: v1alpha1.ManagedNamespaceSpec{
						Namespace: nsReq,
					},
				})).To(BeNil())
			})
		})

		Context("manifest with out quoted string", func() {
			It("should be no issues", func() {
				nsTemplate.NsResources.Resources[0].CustomResource.Manifest = `{"apiVersion":"v1","kind":"ServiceAccount","metadata":{"name":"preprod-sa3","namespace":"second-namespace"}}`
				nsTemplate.NsResources.Resources[0].CustomResource.Manifest = manifest
				Expect(template.ProcessCustomResourceTemplate(context.Background(), v1alpha1.NamespaceTemplate{
					Spec: v1alpha1.NamespaceTemplateSpec{
						NamespaceTemplate: nsTemplate,
					},
				}, &v1alpha1.ManagedNamespace{
					Spec: v1alpha1.ManagedNamespaceSpec{
						Namespace: nsReq,
					},
				})).To(BeNil())
			})
		})

		Context("manifest with exported params", func() {
			It("should be no issues", func() {
				nsTemplate.NsResources.Resources[0].CustomResource.Manifest = `{"apiVersion":"v1","kind":"ServiceAccount","metadata":{"name":"${env}-sa3","namespace":"second-namespace"}}`
				nsTemplate.NsResources.Resources[0].CustomResource.Manifest = manifest
				Expect(template.ProcessCustomResourceTemplate(context.Background(), v1alpha1.NamespaceTemplate{
					Spec: v1alpha1.NamespaceTemplateSpec{
						NamespaceTemplate: nsTemplate,
					},
				}, &v1alpha1.ManagedNamespace{
					Spec: v1alpha1.ManagedNamespaceSpec{
						Namespace: nsReq,
					},
				})).To(BeNil())
			})
		})
	})
})
