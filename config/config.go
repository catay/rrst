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

func (self *ConfigData) SetReposDefaults() {
	for i, r := range self.Repos {
		if r.CacheDir == "" {
			//r.CacheDir = self.Globals.CacheDir
			self.Repos[i].CacheDir = self.Globals.CacheDir
		}
	}
}

func (self *ConfigData) Print() {
	for _, r := range self.Repos {
		fmt.Println("*", r.Name, r.CacheDir)
	}
}
