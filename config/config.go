package config

import (
	"fmt"
	"github.com/catay/rrst/repository"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type ConfigData struct {
	Version int                     `yaml:"version"`
	Globals globals                 `yaml:"globals"`
	Repos   []repository.Repository `yaml:"repositories"`
}

type globals struct {
	CacheDir string `yaml:"cache_dir"`
	ProxyURL string `yaml:"proxy_url"`
}

func NewConfig(configFile string) (c *ConfigData, err error) {
	c = &ConfigData{
		Globals: globals{CacheDir: "/var/cache/rrst"},
	}

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	// file can't be empty
	if len(data) == 0 {
		return nil, fmt.Errorf("%s is empty", configFile)
	}

	if err := yaml.UnmarshalStrict(data, c); err != nil {
		return nil, err
	}

	c.SetReposDefaults()

	return c, err
}

func (c *ConfigData) SetReposDefaults() {
	for i, r := range c.Repos {
		if r.CacheDir == "" {
			//r.CacheDir = c.Globals.CacheDir
			c.Repos[i].CacheDir = c.Globals.CacheDir
		}
	}
}

func (c *ConfigData) Print() {
	for _, r := range c.Repos {
		fmt.Println("*", r.Name, r.CacheDir)
	}
}

// Return a matching repository or nil if not found.
func (c *ConfigData) GetRepoByName(name string) *repository.Repository {
	for _, r := range c.Repos {
		if r.Name == name {
			return &r
		}
	}

	return nil
}
