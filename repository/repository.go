package repository

import (
	"errors"
	"fmt"
	"github.com/catay/rrst/api/suse"
	"github.com/catay/rrst/repomd"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Repository struct {
	Name            string   `yaml:"name"`
	RType           string   `yaml:"type"`
	Vendor          string   `yaml:"vendor"`
	RegCode         string   `yaml:"reg_code"`
	RemoteURI       string   `yaml:"remote_uri"`
	LocalURI        string   `yaml:"local_uri"`
	UpdatePolicy    string   `yaml:"update_policy"`
	UpdateSuffix    string   `yaml:"update_suffix"`
	CacheDir        string   `yaml:"cache_dir"`
	IncludePatterns []string `yaml:"include_patterns"`
	secret          string
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

		self.secret, ok = scc.GetSecretURI(self.RemoteURI)
		if !ok {
			return errors.New(fmt.Sprintf("Secret for url  %v not found", self.RemoteURI))
		}
	}

	fmt.Println("  - Fetch repomd xml file")
	rm := repomd.NewRepoMd(self.RemoteURI, self.secret, self.CacheDir+"/"+self.Name)

	if err := rm.Metadata(); err != nil {
		return err
	}

	fmt.Println("Package count:", rm.PackageCount())

	rpms := rm.Packages()

	for i, p := range rpms {
		ok, err := self.matchPatterns(self.IncludePatterns, p.Loc.Path)
		if err != nil {
			return err
		}
		if ok {
			ok, err := self.packageExistsLocal(p)
			if err != nil {
				return err
			}

			if !ok {
				rpms[i].ToDownload = true
			}

			//		fmt.Println("  - ", rpms[i].Loc.Path, rpms[i].ToDownload)
		}
	}

	for _, p := range rpms {
		if p.ToDownload {
			if err := self.downloadPackage(p); err != nil {
				return err
			}
		}
	}

	fmt.Println("  - Read repomd xml file and get package file location")
	fmt.Println("  - Fetch packages xml file and check hash")
	fmt.Println("  - Read packages xml file and get packages list and check hash")
	fmt.Println("  - Download packages to local path if not existing yet and check hash")

	return nil
}

func (self *Repository) matchPatterns(p []string, s string) (bool, error) {
	if len(self.IncludePatterns) == 0 {
		return true, nil
	}

	for _, p := range p {
		ok, err := regexp.MatchString(p, s)
		if err != nil {
			return false, err
		}

		if ok {
			return true, nil
		}
	}
	return false, nil
}

func (self *Repository) packageExistsLocal(p repomd.RpmPackage) (bool, error) {

	fullPath := self.LocalURI + "/" + p.Loc.Path

	_, err := os.Open(fullPath)

	if err == nil {
		return true, err
	} else {
		if os.IsNotExist(err) {
			return false, nil
		}
	}

	return false, err
}

func (self *Repository) downloadPackage(p repomd.RpmPackage) error {

	rpmDir := self.LocalURI + "/" + filepath.Dir(p.Loc.Path)
	remoteRpmPath := self.RemoteURI + "/" + p.Loc.Path + "?" + self.secret
	localRpmPath := self.LocalURI + "/" + p.Loc.Path

	// create dir
	if err := os.MkdirAll(rpmDir, 0755); err != nil {
		return err
	}

	// create file

	f, err := os.Create(localRpmPath)
	defer f.Close()

	if err != nil {
		return err
	}

	// download

	client := &http.Client{}

	resp, err := client.Get(remoteRpmPath)

	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("HTTP error %v ", resp.StatusCode))
	}

	fmt.Printf("* %s ... ", p.Loc.Path)

	buf := make([]byte, 0, 1)
	var nbytes int64

	for {
		n, err := resp.Body.Read(buf[:cap(buf)])
		buf = buf[:n]
		nbytes += int64(n)
		f.Write(buf)

		if err == io.EOF {
			fmt.Println("successfull")
			break
		}
	}

	return nil

}
