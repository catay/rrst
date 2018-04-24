package repository

import (
	"errors"
	"fmt"
	"github.com/catay/rrst/api/suse"
	"github.com/catay/rrst/repomd"
	"github.com/catay/rrst/util/file"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	tmpSuffix = ".filepart"
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
	topLevelDir     string
	metadata        *repomd.Repomd
}

func NewRepository(name string) (r *Repository) {
	r = &Repository{
		Name: name,
	}

	return
}

func (self *Repository) GetUpdatePolicy() string {
	// return the policy string and set default to merge
	switch self.UpdatePolicy {
	case "stage":
		return self.UpdatePolicy
	default:
		return "merge"
	}
}

func (self *Repository) GetRegCode() (string, bool) {

	if strings.HasPrefix(self.RegCode, "${") && strings.HasSuffix(self.RegCode, "}") {
		key := strings.TrimPrefix(self.RegCode, "${")
		key = strings.TrimSuffix(key, "}")
		return os.LookupEnv(key)
	}

	return self.RegCode, true
}

func (self *Repository) Clean() error {
	rm, err := repomd.NewRepoMd(self.RemoteURI, self.secret, self.CacheDir+"/"+self.Name)
	if err != nil {
		return err
	}

	fmt.Printf("  * %s\n", self.Name)

	if err := rm.Clean(); err != nil {
		return err
	}

	return nil
}

func (self *Repository) Sync() error {

	var err error = nil

	fmt.Printf("  * %s\n", self.Name)

	if self.Vendor == "SUSE" {
		//fmt.Println("  - Fetch SUSE products json if older then x hours")

		regCode, ok := self.GetRegCode()
		if !ok {
			return errors.New(fmt.Sprintf("Environment variable %v not set", self.RegCode))
		}

		scc := suse.NewSCCApi(regCode, self.CacheDir)
		if err := scc.FetchProductsJson(); err != nil {
			return err
		}

		//fmt.Println("  - Get secret hash for give URL repo")

		self.secret, ok = scc.GetSecretURI(self.RemoteURI)
		if !ok {
			return errors.New(fmt.Sprintf("Secret for url  %v not found", self.RemoteURI))
		}
	}

	//fmt.Println("  - Fetch repomd xml file")
	self.metadata, err = repomd.NewRepoMd(self.RemoteURI, self.secret, self.CacheDir+"/"+self.Name)
	if err != nil {
		return err
	}

	if err := self.metadata.Metadata(); err != nil {
		return err
	}

	//fmt.Println("Package count:", self.metadata.PackageCount())

	// given a repository

	// when update policy is set to stage
	// and main dir does not exist
	// then topLevelDir should be set to 'main'

	// when update policy is set to stage
	// and main dir does exist
	// then topLevelDir should be set to 'current timestamp'

	// when update policy is set to merge
	// and main dir does not exist
	// then topLevelDir should be set to 'main'

	// when update policy is set to merge
	// and main dir does exist
	// then topLevelDir should be set to 'main'

	self.topLevelDir = self.LocalURI + "/" + "main"

	if self.GetUpdatePolicy() == "stage" {
		if file.IsDirectory(self.topLevelDir) {
			t := time.Now()
			timeStamp := fmt.Sprintf("%02d%02d%02d-%02d%02d%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
			self.topLevelDir = self.LocalURI + "/" + timeStamp
		} else {
			self.topLevelDir = self.LocalURI + "/" + "main"
		}

	}

	// create toplevel dir
	if err := os.MkdirAll(self.LocalURI, 0755); err != nil {
		return err
	}

	if err := self.markPackagesForDownload(); err != nil {
		return err
	}

	rpms := self.metadata.Packages()

	//fmt.Println("DEBUG - POLICY - ", self.GetUpdatePolicy())
	//fmt.Println("DEBUG - TOPDIR - ", self.topLevelDir)

	for _, p := range rpms {
		//fmt.Printf("package: %v  download: %v\n", p.Name, p.ToDownload)
		if p.ToDownload {
			if err := self.downloadPackage(p); err != nil {
				return err
			}
		}
	}

	//fmt.Println("  - Read repomd xml file and get package file location")
	//fmt.Println("  - Fetch packages xml file and check hash")
	//fmt.Println("  - Read packages xml file and get packages list and check hash")
	//fmt.Println("  - Download packages to local path if not existing yet and check hash")

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

func (self *Repository) markPackagesForDownload() error {

	var localPackages []string

	// build list of local rpm filepaths and store it in localPackages
	err := filepath.Walk(self.LocalURI, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Mode().IsRegular() && strings.HasSuffix(path, "rpm") {
			localPackages = append(localPackages, path)
		}

		if info.Mode().IsRegular() && strings.HasSuffix(path, tmpSuffix) {
			localPackages = append(localPackages, path)
		}

		return nil
	})

	if err != nil {
		return err
	}

	for i, p := range self.metadata.Packages() {
		ok, err := self.matchPatterns(self.IncludePatterns, p.Loc.Path)
		if err != nil {
			return err
		}

		if ok {
			//fmt.Printf("DEBUG MARK p: %v\n", p.Loc.Path)
			self.metadata.PrimaryData.Package[i].ToDownload = true
			//self.metadata.Packages()[i].ToDownload = true
			//p.ToDownload = true

			for _, lp := range localPackages {
				if strings.HasSuffix(lp, p.Loc.Path) {
					//p.ToDownload = false
					self.metadata.PrimaryData.Package[i].ToDownload = false
					//	self.metadata.Packages()[i].ToDownload = false
				}

				// store the local path in case the package was not complelty downloaded
				if strings.HasSuffix(strings.TrimSuffix(lp, tmpSuffix), p.Loc.Path) {
					self.metadata.PrimaryData.Package[i].LocalPath = strings.TrimSuffix(lp, p.Loc.Path+tmpSuffix)
				}
			}
		}
	}

	return nil
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

	rpmDir := self.topLevelDir + "/" + filepath.Dir(p.Loc.Path)
	remoteRpmPath := self.RemoteURI + "/" + p.Loc.Path + "?" + self.secret
	localRpmPath := self.topLevelDir + "/" + p.Loc.Path + tmpSuffix

	if p.LocalPath != "" {
		localRpmPath = p.LocalPath + "/" + p.Loc.Path + tmpSuffix
	}

	// create dir
	if err := os.MkdirAll(rpmDir, 0755); err != nil {
		return err
	}

	// create a new file if it doesn't exist, if file exists open in append mode
	f, err := os.OpenFile(localRpmPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	// get current file size
	stat, err := f.Stat()
	if err != nil {
		return err
	}

	initialFileSize := stat.Size()

	// download the file
	client := &http.Client{}

	resp, err := client.Get(remoteRpmPath)

	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("HTTP error %v ", resp.StatusCode))
	}

	fmt.Printf("    >> %-70s ... ", p.Loc.Path)

	buf := make([]byte, 0, 1)
	var nbytes int64

	for {
		n, err := resp.Body.Read(buf[:cap(buf)])
		buf = buf[:n]
		nbytes += int64(n)

		// only start appending when buffer is greater than initial file size
		if nbytes > initialFileSize {
			f.Write(buf)
		}

		if err == io.EOF {
			fmt.Println("done")
			break
		}
	}

	resp.Body.Close()
	f.Close()

	// rename the file by removing the filepart suffix
	if err := os.Rename(localRpmPath, strings.TrimSuffix(localRpmPath, tmpSuffix)); err != nil {
		return err
	}

	return nil

}
