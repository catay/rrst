package file_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/catay/rrst/util/file"
)

// Feature: file package
// Scenario: testing IsRegularFile(name string) function

// Given a function IsRegularFile(name string)
// When passing an existing regular file as argument
// Then the function should return true

// When passing a non-existing regular file as argument
// Then the function should return false

// When passing a existing directory as argument
// Then the function should return false

var _ = Describe("File package: ", func() {
	var (
		existingDirName     string
		existingFileName    string
		nonExistingDirName  string
		nonExistingFileName string
	)

	BeforeEach(func() {
		existingDirName = "testdata/this_is_a_directory"
		existingFileName = existingDirName + "/this_is_a_regular_file.txt"
		nonExistingDirName = "testdata/this_is_a_not_existing_directory"
		nonExistingFileName = nonExistingDirName + "/this_is_a_not_existing_regular_file.txt"
	})

	Describe("Given a function IsRegularFile(name string)", func() {
		Context("when passing an existing regular file as argument", func() {
			It("should return true", func() {
				v := IsRegularFile(existingFileName)
				Expect(v).To(BeTrue())
			})
		})

		Context("when passing a non-existing regular file as argument", func() {
			It("should return false", func() {
				v := IsRegularFile(nonExistingFileName)
				Expect(v).To(BeFalse())
			})
		})

		Context("when passing an existing directory as argument", func() {
			It("should return false", func() {
				v := IsRegularFile(existingDirName)
				Expect(v).To(BeFalse())
			})
		})
	})

	Describe("Given a function IsDirectory(name string)", func() {
		Context("when passing an existing directory as argument", func() {
			It("should return true", func() {
				v := IsDirectory(existingDirName)
				Expect(v).To(BeTrue())
			})
		})

		Context("when passing a non-existing directory as argument", func() {
			It("should return false", func() {
				v := IsDirectory(nonExistingDirName)
				Expect(v).To(BeFalse())
			})
		})

		Context("when passing an existing regular file as argument", func() {
			It("should return false", func() {
				v := IsDirectory(existingFileName)
				Expect(v).To(BeFalse())
			})
		})
	})
})
