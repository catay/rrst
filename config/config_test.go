package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/catay/rrst/config"
)

var _ = Describe("Config", func() {

	var configFile string
	var config *ConfigData
	var err error

	BeforeEach(func() {
		configFile = "testdata/config_valid.yaml"
	})

	JustBeforeEach(func() {
		config, err = NewConfig(configFile)
	})

	Describe("YAML configuration file", func() {
		Context("when initializing a valid YAML configuration file", func() {

			It("has the global cache dir set", func() {
				Expect(config.Globals.CacheDir).To(Equal("/var/tmp/rrst/cache"))
			})

			It("has repositories set", func() {
				Expect(len(config.Repos)).To(Equal(2))
			})

			It("has the cache dir of repositories set", func() {
				for _, r := range config.Repos {
					Expect(r.CacheDir).To(Equal(config.Globals.CacheDir))
				}
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
