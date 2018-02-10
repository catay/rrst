package config

import (
	"os"
	"strings"
)

type Repository struct {
	Name         string `yaml:"name"`
	RType        string `yaml:"type"`
	Vendor       string `yaml:"vendor"`
	RegCode      string `yaml:"reg_code"`
	RemoteURI    string `yaml:"remote_uri"`
	LocalURI     string `yaml:"local_uri"`
	UpdatePolicy string `yaml:"update_policy"`
	UpdateSuffix string `yaml:"update_suffix"`
}

func NewRepository(name string) (r *Repository) {
	r = &Repository{
		Name: name,
	}

	return
}

func (self *Repository) GetRegCode() (string, bool) {

	if strings.HasPrefix(self.RegCode, "${") && strings.HasSuffix(self.RegCode, "}") {
		key := strings.TrimPrefix(self.RegCode, "${")
		key = strings.TrimSuffix(key, "}")
		return os.LookupEnv(key)
	}

	return self.RegCode, true
}
