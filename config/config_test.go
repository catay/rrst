package config_test

import (
	. "github.com/catay/rrst/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {

	var configFile string
	var config *Config
	var err error

	BeforeEach(func() {
		configFile = "testdata/config_valid.yaml"
	})

	JustBeforeEach(func() {
		config, err = NewConfig(configFile)
	})

	Describe("YAML configuration file", func() {
		Context("when initializing a valid YAML configuration file", func() {

			It("has the global content path set", func() {
				Expect(config.GlobalConfig.ContentPath).To(Equal("/var/tmp/rrst"))
			})

			It("has the max tags to keep set", func() {
				Expect(config.GlobalConfig.MaxRevisionsToKeep).To(Equal(10))
			})

			It("has repositories configured", func() {
				Expect(len(config.RepoConfigs)).To(Equal(2))
			})

			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

		})

		Context("when initializing an empty YAML configuration file", func() {
			BeforeEach(func() {
				configFile = "testdata/config_empty.yaml"
			})

			It("should error", func() {
				Expect(err).To(HaveOccurred())
				Expect(config).Should(BeNil())
			})
		})

		Context("when initializing an not valid YAML configuration file", func() {
			BeforeEach(func() {
				configFile = "testdata/config_not_valid.yaml"
			})

			It("should error", func() {
				Expect(err).To(HaveOccurred())
				Expect(config).Should(BeNil())
			})
		})

		Context("when a valid YAML configuration file is missing", func() {
			BeforeEach(func() {
				configFile = "testdata/config_not_exists.yaml"
			})

			It("should error", func() {
				Expect(err).To(HaveOccurred())
				Expect(config).Should(BeNil())
			})
		})

	})

})
