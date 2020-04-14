package utils_test

import (
	"github.com/keikoproj/manager/internal/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("internal.utils.utils test cases", func() {
	Describe("BoolValue test case ", func() {

		Context("Successful use case with true", func() {
			It("bool should be true", func() {
				Expect(utils.BoolValue("true")).To(BeTrue())
			})
		})

		Context("Successful use case with false", func() {
			It("bool should be false", func() {
				Expect(utils.BoolValue("false")).To(BeFalse())
			})
		})

		Context("Successful use case with empty string", func() {
			It("bool should be false", func() {
				Expect(utils.BoolValue("")).To(BeFalse())
			})
		})

		Context("Successful use case with case sensitive string True", func() {
			It("bool should be true", func() {
				Expect(utils.BoolValue("True")).To(BeTrue())
			})
		})

		Context("Successful use case with case sensitive string FALSE", func() {
			It("bool should be false", func() {
				Expect(utils.BoolValue("FALSE")).To(BeFalse())
			})
		})
	})

	Describe("SanitizeString() test cases", func() {
		Context("valid string with no changes needed", func() {
			It("should return the same string", func() {
				Expect(utils.SanitizeName("somethinggood")).To(Equal("somethinggood"))
			})
		})
		Context("valid string with replacement", func() {
			It("should return the string without any dots", func() {
				Expect(utils.SanitizeName("dev-patterns.manager-usw2.ppd-idev")).To(Equal("dev-patterns-manager-usw2-ppd-idev"))
			})
		})
	})

	Describe("ContainsString() test cases", func() {
		Context("valid comparision", func() {
			It("should be true", func() {
				Expect(utils.ContainsString([]string{"iamrole.finalizers.iammanager.keikoproj.io", "iamrole.finalizers2.iammanager.keikoproj.io"}, "iamrole.finalizers.iammanager.keikoproj.io")).To(BeTrue())
			})
		})
		Context("different string comparision", func() {
			It("should return false", func() {
				Expect(utils.ContainsString([]string{"iamrole.finalizers.iammanager.keikoproj.io", "iamrole.finalizers2.iammanager.keikoproj.io"}, "iamrole-iammanager.keikoproj.io")).To(BeFalse())
			})
		})
	})

	Describe("RemoveString() test cases", func() {
		var emptySlice []string
		Context("should remove one value", func() {
			It("should be equal to the remaining string", func() {
				Expect(utils.RemoveString([]string{"iamrole.finalizers.iammanager.keikoproj.io", "iamrole.finalizers2.iammanager.keikoproj.io"}, "iamrole.finalizers.iammanager.keikoproj.io")).To(Equal([]string{"iamrole.finalizers2.iammanager.keikoproj.io"}))
			})
		})
		Context("empty slice with remove usecase", func() {
			It("should just return the empty slice", func() {
				Expect(utils.RemoveString([]string{}, "iamrole.finalizers.iammanager.keikoproj.io")).To(Equal(emptySlice))
			})
		})
		Context("empty slice with remove usecase", func() {
			It("should just return the empty slice", func() {
				Expect(utils.RemoveString([]string{}, "iamrole.finalizers.iammanager.keikoproj.io")).To(Equal(emptySlice))
			})
		})
		Context("empty the slice by removing one string", func() {
			It("should just return the empty slice", func() {
				Expect(utils.RemoveString([]string{"iamrole.finalizers.iammanager.keikoproj.io"}, "iamrole.finalizers.iammanager.keikoproj.io")).To(Equal(emptySlice))
			})
		})

		Context("trying to remove the value which doesn't exists", func() {
			It("should just return the original slice", func() {
				Expect(utils.RemoveString([]string{"iamrole.finalizers.iammanager.keikoproj.io"}, "iamrole.finalizers2.iammanager.keikoproj.io")).To(Equal([]string{"iamrole.finalizers.iammanager.keikoproj.io"}))
			})
		})
	})

})
