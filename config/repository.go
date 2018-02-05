package config

type repository struct {
	Name         string `yaml:"name"`
	RType        string `yaml:"type"`
	Vendor       string `yaml:"vendor"`
	RegCode      string `yaml:"reg_code"`
	RemoteURI    string `yaml:"remote_uri"`
	LocalURI     string `yaml:"local_uri"`
	UpdatePolicy string `yaml:"update_policy"`
	UpdateSuffix string `yaml:"update_suffix"`
}
