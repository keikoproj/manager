package controllers

import (
	"context"
	"errors"
	"github.com/keikoproj/manager/pkg/grpc/proto/namespace"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ManagedNamespaceController", func() {
	Describe("shouldProceed first iteration ", func() {

		statusMap := make(map[string]ResourceStatus)

		Context("Service Account 1 shouldn't be created in first iteration", func() {
			It("should be false", func() {
				Expect(shouldProceed(context.Background(), statusMap, &namespace.Resource{
					Name:      "local_service_account1",
					DependsOn: "local_role",
				}, false)).To(BeFalse())
			})
		})
		Context("Service Account 2 should be created in first iteration", func() {
			It("should be true", func() {
				Expect(shouldProceed(context.Background(), statusMap, &namespace.Resource{
					Name: "local_service_account2",
				}, false)).To(BeTrue())
			})
		})
		Context("local role should be created in first iteration", func() {
			It("should be true", func() {
				Expect(shouldProceed(context.Background(), statusMap, &namespace.Resource{
					Name: "local_role",
				}, false)).To(BeTrue())
			})
		})
		Context("local role binding shouldn't be created in first iteration", func() {
			It("should be false", func() {
				Expect(shouldProceed(context.Background(), statusMap, &namespace.Resource{
					Name:      "local_role_binding",
					DependsOn: "local_service_account1",
				}, false)).To(BeFalse())
			})
		})
	})

	Describe("shouldProceed with conditions considering dependencies are done- 1st level ", func() {

		statusMap := make(map[string]ResourceStatus)

		statusMap["local_service_account1"] = ResourceStatus{
			Name:      "local_service_account1",
			DependsOn: "local_role",
		}
		statusMap["local_service_account2"] = ResourceStatus{
			Name: "local_service_account2",
			Done: true,
		}
		statusMap["local_role"] = ResourceStatus{
			Name: "local_role",
			Done: true,
		}
		statusMap["local_role_binding"] = ResourceStatus{
			Name:      "local_role_binding",
			DependsOn: "local_service_account1",
		}

		Context("Service Account 1 should be created as local_role is already exists", func() {
			It("should be true as local_role is already created", func() {
				Expect(shouldProceed(context.Background(), statusMap, &namespace.Resource{
					Name:      "local_service_account1",
					DependsOn: "local_role",
				}, false)).To(BeTrue())
			})
		})
		Context("Service Account 2 shouldn't proceed as it already is done", func() {
			It("should be false", func() {
				Expect(shouldProceed(context.Background(), statusMap, &namespace.Resource{
					Name: "local_service_account2",
				}, false)).To(BeFalse())
			})
		})
		Context("local role shouldn't proceed as it already is done", func() {
			It("should be false", func() {
				Expect(shouldProceed(context.Background(), statusMap, &namespace.Resource{
					Name: "local_role",
				}, false)).To(BeFalse())
			})
		})
		Context("local role binding shouldn't be created in since service account1 is not done yet", func() {
			It("should be false", func() {
				Expect(shouldProceed(context.Background(), statusMap, &namespace.Resource{
					Name:      "local_role_binding",
					DependsOn: "local_service_account1",
				}, false)).To(BeFalse())
			})
		})
	})

	Describe("shouldProceed with conditions considering dependencies are done- 2nd level ", func() {

		statusMap := make(map[string]ResourceStatus)

		statusMap["local_service_account1"] = ResourceStatus{
			Name:      "local_service_account1",
			DependsOn: "local_role",
			Done:      true,
		}
		statusMap["local_service_account2"] = ResourceStatus{
			Name: "local_service_account2",
			Done: true,
		}
		statusMap["local_role"] = ResourceStatus{
			Name: "local_role",
			Done: true,
		}
		statusMap["local_role_binding"] = ResourceStatus{
			Name:      "local_role_binding",
			DependsOn: "local_service_account1",
		}

		Context("Service Account 1 shouldn't proceed as it already is done", func() {
			It("should be false", func() {
				Expect(shouldProceed(context.Background(), statusMap, &namespace.Resource{
					Name:      "local_service_account1",
					DependsOn: "local_role",
				}, false)).To(BeFalse())
			})
		})
		Context("Service Account 2 shouldn't proceed as it already is done", func() {
			It("should be false", func() {
				Expect(shouldProceed(context.Background(), statusMap, &namespace.Resource{
					Name: "local_service_account2",
				}, false)).To(BeFalse())
			})
		})
		Context("local role shouldn't proceed as it already is done", func() {
			It("should be false", func() {
				Expect(shouldProceed(context.Background(), statusMap, &namespace.Resource{
					Name: "local_role",
				}, false)).To(BeFalse())
			})
		})
		Context("local role binding should be created now as its dependencies are done", func() {
			It("should be true", func() {
				Expect(shouldProceed(context.Background(), statusMap, &namespace.Resource{
					Name:      "local_role_binding",
					DependsOn: "local_service_account1",
				}, false)).To(BeTrue())
			})
		})
	})

	Describe("shouldProceed with conditions when dependsOn and error ", func() {

		statusMap := make(map[string]ResourceStatus)

		statusMap["local_service_account1"] = ResourceStatus{
			Name:      "local_service_account1",
			DependsOn: "local_role",
			Error:     errors.New("something"),
		}

		Context("Service Account 1 shouldn't proceed as there is an error", func() {
			It("should be false", func() {
				Expect(shouldProceed(context.Background(), statusMap, &namespace.Resource{
					Name: "local_service_account1",
				}, false)).To(BeFalse())
			})
		})

	})

	Describe("shouldProceed with just an error ", func() {

		statusMap := make(map[string]ResourceStatus)

		statusMap["local_service_account1"] = ResourceStatus{
			Name:  "local_service_account1",
			Error: errors.New("something"),
		}

		Context("Service Account 1 shouldn't proceed as there is an error", func() {
			It("should be false", func() {
				Expect(shouldProceed(context.Background(), statusMap, &namespace.Resource{
					Name: "local_service_account1",
				}, false)).To(BeFalse())
			})
		})

	})

	Describe("createOnly use case with firstTime ", func() {

		statusMap := make(map[string]ResourceStatus)

		statusMap["local_service_account1"] = ResourceStatus{
			Name: "local_service_account1",
		}

		Context("successful use case", func() {
			It("should be true", func() {
				Expect(shouldProceed(context.Background(), statusMap, &namespace.Resource{
					Name:       "local_service_account1",
					CreateOnly: "true",
				}, true)).To(BeTrue())
			})
		})
		Context("Should not proceed as it is not first time", func() {
			It("should be true", func() {
				Expect(shouldProceed(context.Background(), statusMap, &namespace.Resource{
					Name:       "local_service_account1",
					CreateOnly: "true",
				}, false)).To(BeFalse())
			})
		})

		Context("dependency use case", func() {
			It("should be true", func() {
				Expect(shouldProceed(context.Background(), statusMap, &namespace.Resource{
					Name:       "local_service_account1",
					CreateOnly: "true",
					DependsOn:  "something",
				}, false)).To(BeFalse())
			})
		})
	})

})
