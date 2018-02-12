package config

import (
	"errors"
	"fmt"
	"github.com/catay/rrst/api/suse"
	"github.com/catay/rrst/repomd"
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
	CacheDir     string `yaml:"cache_dir"`
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

func (self *Repository) Sync() error {

	var secret string

	if self.Vendor == "SUSE" {
		fmt.Println("  - Fetch SUSE products json if older then x hours")

		regCode, ok := self.GetRegCode()
		if !ok {
			return errors.New(fmt.Sprintf("Environment variable %v not set", self.RegCode))
		}

		scc := suse.NewSCCApi(regCode, self.CacheDir)
		if err := scc.FetchProductsJson(); err != nil {
			return err
		}

		fmt.Println("  - Get secret hash for give URL repo")

		secret, ok = scc.GetSecretURI(self.RemoteURI)
		if !ok {
			return errors.New(fmt.Sprintf("Secret for url  %v not found", self.RemoteURI))
		}
	}

	fmt.Println("  - Fetch repomd xml file")
	rm := repomd.NewRepoMd(self.RemoteURI, secret, self.CacheDir+"/"+self.Name)

	if err := rm.Metadata(); err != nil {
		return err
	}

	fmt.Println("  - Read repomd xml file and get package file location")
	fmt.Println("  - Fetch packages xml file and check hash")
	fmt.Println("  - Read packages xml file and get packages list and check hash")
	fmt.Println("  - Download packages to local path if not existing yet and check hash")

	return nil
}
