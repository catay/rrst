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
	Name      string `yaml:"name"`
	RType     string `yaml:"type"`
	Vendor    string `yaml:"Vendor"`
	RegCode   string `yaml:"reg_code"`
	RemoteURI string `yaml:"remote_uri"`
	LocalURI  string `yaml:"local_uri"`
}

func New(configFile string) (c *ConfigData, err error) {
	c = new(ConfigData)

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, c); err != nil {
		return nil, err
	}

	return c, err
}

func (self *ConfigData) Print() {
	for _, r := range self.Repos {
		fmt.Println("*", r.Name)
	}
}
