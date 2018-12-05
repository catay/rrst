package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

// Default configuration values.
const (
	DefaultContentPath            = "~/.rrst/content"
	DefaultMaxTagsToKeep          = 50
	DefaultContentFilesPathSuffix = "files"
	DefaultContentMDPathSuffix    = "metadata"
	DefaultContentTmpPathSuffix   = "tmp"
	ContentPathEnv                = "RRST_CONTENT_PATH"
)

// Config is the top-level configuration for rrst.
type Config struct {
	Version      string              `yaml:"version"`
	GlobalConfig GlobalConfig        `yaml:"global"`
	RepoConfigs  []*RepositoryConfig `yaml:"repositories"`
}

// GlobalConfig contains the global configuration settings.
type GlobalConfig struct {
	ContentPath      string `yaml:"content_path"`
	MaxTagsToKeep    int    `yaml:"max_tags_to_keep"`
	ContentFilesPath string
	ContentMDPath    string
	ContentTmpPath   string
}

// RepositoryConfig contains the per repository configuration settings.
type RepositoryConfig struct {
	Id                int      `yaml:"id"`
	Name              string   `yaml:"name"`
	RType             string   `yaml:"type"`
	Vendor            string   `yaml:"vendor"`
	RegCode           string   `yaml:"reg_code"`
	RemoteURI         string   `yaml:"remote_uri"`
	ContentSuffixPath string   `yaml:"content_suffix_path"`
	IncludePatterns   []string `yaml:"include_patterns"`
	MaxTagsToKeep     int      `yaml:"max_tags_to_keep"`
	Enabled           bool     `yaml:"enabled"`
	ContentPath       string
	ContentFilesPath  string
	ContentMDPath     string
	ContentTmpPath    string
}

// NewConfig loads the configuration from a YAML file and returns it.
// The config will be nil when an error is encountered.
func NewConfig(configFile string) (c *Config, err error) {
	// Set the configuration default values.
	c = &Config{
		GlobalConfig: GlobalConfig{
			ContentPath:   DefaultContentPath,
			MaxTagsToKeep: DefaultMaxTagsToKeep,
		},
	}

	// Set content path from environment variable when present.
	if v := os.Getenv(ContentPathEnv); v != "" {
		c.GlobalConfig.ContentPath = v
	}

	// Load the configuration values from the YAML config file.
	if err := c.LoadFromYAMLFile(configFile); err != nil {
		return nil, fmt.Errorf("error loading config: %s", err)
	}

	// Set content path default sub directories
	c.GlobalConfig.ContentFilesPath = c.GlobalConfig.ContentPath + "/" + DefaultContentFilesPathSuffix
	c.GlobalConfig.ContentMDPath = c.GlobalConfig.ContentPath + "/" + DefaultContentMDPathSuffix
	c.GlobalConfig.ContentTmpPath = c.GlobalConfig.ContentPath + "/" + DefaultContentTmpPathSuffix

	// Set repository configuration defaults when not set after
	// loading the YAML file.
	c.SetRepositoryConfigDefaults()

	return c, nil
}

// Load a configuration from a YAML file and set the config values.
func (c *Config) LoadFromYAMLFile(configFile string) (err error) {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	// file can't be empty
	if len(data) == 0 {
		return fmt.Errorf("open %s: file is empty", configFile)
	}

	if err := yaml.UnmarshalStrict(data, c); err != nil {
		return err
	}

	return nil
}

// Set the repository configuration defaults when not set.
func (c *Config) SetRepositoryConfigDefaults() {
	// Loop over all repo configs and set defaults when not
	// initialized.
	for i, r := range c.RepoConfigs {
		if r.ContentPath == "" {
			c.RepoConfigs[i].ContentPath = c.GlobalConfig.ContentPath
		}

		if r.MaxTagsToKeep == 0 {
			c.RepoConfigs[i].MaxTagsToKeep = c.GlobalConfig.MaxTagsToKeep
		}

		c.RepoConfigs[i].ContentFilesPath = c.RepoConfigs[i].ContentPath + "/" + DefaultContentFilesPathSuffix
		c.RepoConfigs[i].ContentMDPath = c.RepoConfigs[i].ContentPath + "/" + DefaultContentMDPathSuffix
		c.RepoConfigs[i].ContentTmpPath = c.RepoConfigs[i].ContentPath + "/" + DefaultContentTmpPathSuffix
	}
}
