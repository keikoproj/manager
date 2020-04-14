package validation_test

import (
	"context"
	"github.com/keikoproj/manager/pkg/grpc/proto/namespace"
	"github.com/keikoproj/manager/pkg/validation"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Validation in namespace file", func() {
	Describe("Resource names must be unique ", func() {
		res1 := &namespace.Resource{
			Name: "name1",
		}
		res2 := &namespace.Resource{
			Name: "name2",
		}
		res3 := &namespace.Resource{
			Name: "name3",
		}

		Context("Successful use case", func() {
			It("Error should be nil", func() {
				Expect(validation.ValidateTemplate(context.Background(), &namespace.NamespaceResources{
					Namespace: &v1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name: "Namespace1",
						},
					},
					Resources: []*namespace.Resource{res1, res2, res3},
				})).To(BeNil())
			})
		})

		Context("Duplicate use case", func() {
			It("Error should NOT be nil", func() {
				Expect(validation.ValidateTemplate(context.Background(), &namespace.NamespaceResources{
					Namespace: &v1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name: "Namespace1",
						},
					},
					Resources: []*namespace.Resource{res1, res2, res3, res1},
				})).ToNot(BeNil())
			})
		})

		Context("Duplicate use case with namespace name", func() {
			It("Error should NOT be nil", func() {
				Expect(validation.ValidateTemplate(context.Background(), &namespace.NamespaceResources{
					Namespace: &v1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name: "name1",
						},
					},
					Resources: []*namespace.Resource{res1, res2, res3},
				})).ToNot(BeNil())
			})
		})

		Context("Duplicate use case case sensitive", func() {
			It("Error should NOT be nil", func() {
				Expect(validation.ValidateTemplate(context.Background(), &namespace.NamespaceResources{
					Namespace: &v1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name: "Name1",
						},
					},
					Resources: []*namespace.Resource{res1, res2, res3},
				})).To(BeNil())
			})
		})
	})

	Describe("DependsOn value belongs to one of the resource in the same template ", func() {
		res1 := &namespace.Resource{
			Name:      "name1",
			DependsOn: "name3",
		}
		res2 := &namespace.Resource{
			Name:      "name2",
			DependsOn: "name5",
		}
		res3 := &namespace.Resource{
			Name: "name3",
		}

		Context("Successful use case", func() {
			It("Error should be nil", func() {
				Expect(validation.ValidateTemplate(context.Background(), &namespace.NamespaceResources{
					Namespace: &v1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name: "Namespace1",
						},
					},
					Resources: []*namespace.Resource{res1, res3},
				})).To(BeNil())
			})
		})

		Context("Name doesn't exist", func() {
			It("Error should NOT be nil", func() {
				Expect(validation.ValidateTemplate(context.Background(), &namespace.NamespaceResources{
					Namespace: &v1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name: "Namespace1",
						},
					},
					Resources: []*namespace.Resource{res1, res2},
				})).ToNot(BeNil())
			})
		})

		Context("Only 1 resource with dependency", func() {
			It("Error should be nil", func() {
				Expect(validation.ValidateTemplate(context.Background(), &namespace.NamespaceResources{
					Namespace: &v1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name: "Namespace1",
						},
					},
					Resources: []*namespace.Resource{res2},
				})).ToNot(BeNil())
			})
		})

	})

	Describe("DependsOn Circular Dependency ", func() {
		res1 := &namespace.Resource{
			Name:      "name1",
			DependsOn: "name3",
		}
		res2 := &namespace.Resource{
			Name:      "name2",
			DependsOn: "name1",
		}
		res3 := &namespace.Resource{
			Name: "name3",
		}

		res4 := &namespace.Resource{
			Name:      "name4",
			DependsOn: "name5",
		}
		res5 := &namespace.Resource{
			Name:      "name5",
			DependsOn: "name4",
		}

		res6 := &namespace.Resource{
			Name:      "name6",
			DependsOn: "name7",
		}
		res7 := &namespace.Resource{
			Name:      "name7",
			DependsOn: "name8",
		}
		res8 := &namespace.Resource{
			Name:      "name8",
			DependsOn: "name6",
		}

		Context("Successful use case", func() {
			It("Error should be nil", func() {
				Expect(validation.ValidateTemplate(context.Background(), &namespace.NamespaceResources{
					Namespace: &v1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name: "Namespace1",
						},
					},
					Resources: []*namespace.Resource{res1, res2, res3},
				})).To(BeNil())
			})
		})

		Context("res4 and res5 pointing to each other", func() {
			It("Error should NOT be nil", func() {
				Expect(validation.ValidateTemplate(context.Background(), &namespace.NamespaceResources{
					Namespace: &v1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name: "Namespace1",
						},
					},
					Resources: []*namespace.Resource{res1, res2, res3, res4, res5},
				})).ToNot(BeNil())
			})
		})

		Context("res4 and res5 pointing to each other and only 2 resources", func() {
			It("Error should NOT be nil", func() {
				Expect(validation.ValidateTemplate(context.Background(), &namespace.NamespaceResources{
					Namespace: &v1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name: "Namespace1",
						},
					},
					Resources: []*namespace.Resource{res4, res5},
				})).ToNot(BeNil())
			})
		})
		Context("Only 1 resource without dependency", func() {
			It("Error should be nil", func() {
				Expect(validation.ValidateTemplate(context.Background(), &namespace.NamespaceResources{
					Namespace: &v1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name: "Namespace1",
						},
					},
					Resources: []*namespace.Resource{res3},
				})).To(BeNil())
			})
		})

		Context("3 resources with dependency triangle dependency", func() {
			It("Error should NOT be nil", func() {
				Expect(validation.ValidateTemplate(context.Background(), &namespace.NamespaceResources{
					Namespace: &v1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name: "Namespace1",
						},
					},
					Resources: []*namespace.Resource{res6, res7, res8},
				})).ToNot(BeNil())
			})
		})

		Context("3 resources with dependency clean dependency", func() {
			It("Error should be nil", func() {
				Expect(validation.ValidateTemplate(context.Background(), &namespace.NamespaceResources{
					Namespace: &v1.Namespace{
						ObjectMeta: metav1.ObjectMeta{
							Name: "Namespace1",
						},
					},
					Resources: []*namespace.Resource{
						{
							Name:      "local_service_account1",
							DependsOn: "local_role",
						},
						{
							Name: "local_role",
						},
						{
							Name:      "local_role_binding",
							DependsOn: "local_service_account1",
						},
					},
				})).To(BeNil())
			})
		})
	})

})
