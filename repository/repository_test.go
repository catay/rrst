package repository_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/catay/rrst/repository"
	"os"
)

var _ = Describe("Repository", func() {

	var repo *Repository
	//var err error

	BeforeEach(func() {
		repo = NewRepository("SLES-12-3-X86_64-updates")
		//		repo.RegCode = ""
	})

	//	JustBeforeEach(func() {
	//		repo = NewRepository("SLES-12-3-X86_64-updates")
	//	})

	Describe("New Repository", func() {
		Context("when initializing a new repository", func() {
			It("has a repo name set", func() {
				Expect(repo.Name).To(Equal("SLES-12-3-X86_64-updates"))
			})
			It("has no repo type set", func() {
				Expect(repo.RType).To(BeZero())
			})
			It("has no vendor set", func() {
				Expect(repo.Vendor).To(BeZero())
			})
			It("has no reg code set", func() {
				Expect(repo.RegCode).To(BeZero())
			})
			It("has no remote URI set", func() {
				Expect(repo.RemoteURI).To(BeZero())
			})
			It("has no local URI set", func() {
				Expect(repo.LocalURI).To(BeZero())
			})
			It("has no update policy set", func() {
				Expect(repo.UpdatePolicy).To(BeZero())
			})
			It("has no update suffix set", func() {
				Expect(repo.UpdateSuffix).To(BeZero())
			})
		})

		Context("when reg code contains environment variable which is set", func() {
			BeforeEach(func() {
				os.Setenv("SCC_REG_CODE_01", "666666")
				repo.RegCode = "${SCC_REG_CODE_01}"
			})

			It("should have value 666666 and true", func() {
				regCode, ok := repo.GetRegCode()
				Expect(regCode).To(Equal("666666"))
				Expect(ok).To(BeTrue())
			})

		})

		Context("when reg code contains environment variable not set", func() {
			BeforeEach(func() {
				os.Unsetenv("SCC_REG_CODE_01")
				repo.RegCode = "${SCC_REG_CODE_01}"
			})

			It("should have empty value and false", func() {
				regCode, ok := repo.GetRegCode()
				Expect(regCode).To(BeZero())
				Expect(ok).To(BeFalse())
			})
		})

		Context("when reg code is set through string", func() {
			BeforeEach(func() {
				repo.RegCode = "666666"
			})

			It("should have value 666666 and true", func() {
				regCode, ok := repo.GetRegCode()
				Expect(regCode).To(Equal("666666"))
				Expect(ok).To(BeTrue())
			})
		})

		Context("when reg code is empty", func() {

			It("should have empty value and true", func() {
				regCode, ok := repo.GetRegCode()
				Expect(regCode).To(BeZero())
				Expect(ok).To(BeTrue())
			})
		})

	})
})
