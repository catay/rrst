package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type ConfigData struct {
	Globals globals        `yaml:"globals"`
	Repos   []repositories `yaml:"repositories"`
}

type globals struct {
	CacheDir string `yaml:"cache_dir"`
}

type repositories struct {
	Name         string `yaml:"name"`
	RType        string `yaml:"type"`
	Vendor       string `yaml:"vendor"`
	RegCode      string `yaml:"reg_code"`
	RemoteURI    string `yaml:"remote_uri"`
	LocalURI     string `yaml:"local_uri"`
	UpdatePolicy string `yaml:"update_policy"`
	UpdateSuffix string `yaml:"update_suffix"`
}

func NewConfig(configFile string) (c *ConfigData, err error) {
	c = &ConfigData{}

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

	return c, err
}

func (self *ConfigData) Print() {
	for _, r := range self.Repos {
		fmt.Println("*", r.Name)
	}
}
