package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
)

// Default configuration values.
const (
	ValidTagsRegex                = "^[a-z|A-Z|0-9|_]+$"
	DefaultConfigPath             = "/etc/rrst.yaml"
	DefaultServerPort             = "4280"
	DefaultContentPath            = "~/.rrst/content"
	DefaultMaxRevisionsToKeep     = 50
	DefaultContentFilesPathSuffix = "files"
	DefaultContentMDPathSuffix    = "metadata"
	DefaultContentTmpPathSuffix   = "tmp"
	DefaultContentTagsPathSuffix  = "tags"
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
	ContentPath        string      `yaml:"content_path"`
	Providers          []*Provider `yaml:"providers"`
	MaxRevisionsToKeep int         `yaml:"max_tags_to_keep"`
}

// RepositoryConfig contains the per repository configuration settings.
type RepositoryConfig struct {
	Id                 int      `yaml:"id"`
	Name               string   `yaml:"name"`
	RType              string   `yaml:"type"`
	ProviderId         string   `yaml:"provider_id"`
	RegCode            string   `yaml:"reg_code"`
	RemoteURI          string   `yaml:"remote_uri"`
	ContentSuffixPath  string   `yaml:"content_suffix_path"`
	IncludePatterns    []string `yaml:"include_patterns"`
	MaxRevisionsToKeep int      `yaml:"max_tags_to_keep"`
	Enabled            bool     `yaml:"enabled"`
	ContentFilesPath   string
	ContentMDPath      string
	ContentTagsPath    string
	ContentTmpPath     string
	Provider           *Provider
}

// Provider contains provider specific configuration settings.
//
// Currently only SUSE SCC credentials are supported.
// Id has to contain a free to choose name to which a  repo can map too.
// Name should be the provider name, for example SUSE.
// The variables can contain log in credentials. In the case of SUSE SCC
// it needs to contain for example the variable scc_reg_code which should
// contain a SCC registration code.
type Provider struct {
	Id        string `yaml:"id"`
	Name      string `yaml:"provider"`
	Variables []struct {
		Name  string `yaml:"name"`
		Value string `yaml:"value"`
	} `yaml:"variables"`
}

// NewConfig loads the configuration from a YAML file and returns it.
// The config will be nil when an error is encountered.
func NewConfig(configFile string) (c *Config, err error) {
	// Set the configuration default values.
	c = &Config{
		GlobalConfig: GlobalConfig{
			ContentPath:        DefaultContentPath,
			MaxRevisionsToKeep: DefaultMaxRevisionsToKeep,
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

	// Substitute substitution environment variables when present.
	for i, _ := range c.GlobalConfig.Providers {
		c.GlobalConfig.Providers[i].SetEnvVars()
	}

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
	// overrided at repo level.
	for i, r := range c.RepoConfigs {
		if r.MaxRevisionsToKeep == 0 {
			c.RepoConfigs[i].MaxRevisionsToKeep = c.GlobalConfig.MaxRevisionsToKeep
		}

		c.RepoConfigs[i].ContentFilesPath = c.GlobalConfig.ContentPath + "/" + DefaultContentFilesPathSuffix + "/" + r.ContentSuffixPath
		c.RepoConfigs[i].ContentMDPath = c.GlobalConfig.ContentPath + "/" + DefaultContentMDPathSuffix + "/" + r.ContentSuffixPath
		c.RepoConfigs[i].ContentTagsPath = c.GlobalConfig.ContentPath + "/" + DefaultContentTagsPathSuffix + "/" + r.ContentSuffixPath
		c.RepoConfigs[i].ContentTmpPath = c.GlobalConfig.ContentPath + "/" + DefaultContentTmpPathSuffix + "/" + r.ContentSuffixPath

		// reference repo provider id to provider data if present and match found.
		for _, p := range c.GlobalConfig.Providers {
			if r.ProviderId == p.Id {
				c.RepoConfigs[i].Provider = p
			}
		}
	}
}

// SetEnvVars checks if the value of a provider variable references a
// environment variable and does the substitution when present
// If not present the original value is retained.
func (p *Provider) SetEnvVars() {
	for i, v := range p.Variables {
		if IsEnvVar(v.Value) {
			value, ok := EnvVarValue(v.Value)
			if !ok {
				fmt.Printf("config: warning: no env var set for %v\n", v.Name)
			}
			p.Variables[i].Value = value
		}
	}
}

// EnvVarValue returns the value of the environment variable.
// If environment value is not set, the bool will be false.
func EnvVarValue(v string) (string, bool) {
	if strings.HasPrefix(v, "${") && strings.HasSuffix(v, "}") {
		key := strings.TrimPrefix(v, "${")
		key = strings.TrimSuffix(key, "}")
		return os.LookupEnv(key)
	}

	return v, false
}

// IsEnvVar checks if the value is a substitution variable.
// Returns true or false.
func IsEnvVar(v string) bool {
	if strings.HasPrefix(v, "${") && strings.HasSuffix(v, "}") {
		return true
	}
	return false
}
