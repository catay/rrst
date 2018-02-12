package repomd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

type repomd struct {
	Url      string
	Secret   string
	CacheDir string
	//  metadata string = "repomd.xml"
}

func NewRepoMd(url, secret string, cacheDir string) *repomd {
	r := new(repomd)
	r.Url = url
	r.Secret = secret
	r.CacheDir = cacheDir
	return r
}

func (self *repomd) Debug() {
	fmt.Printf("Url: %v\n", self.Url)
	fmt.Printf("Secret: %v\n", self.Secret)
}

func (self *repomd) Metadata() error {
	return self.fetchRepomdFile("/repodata/repomd.xml")
}

func (self *repomd) fetchRepomdFile(fileLocation string) error {

	resp, err := http.Get(self.Url + fileLocation + "?" + self.Secret)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("HTTP error %v ", resp.StatusCode))
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// check if dir exists, if not create it
	if _, err := os.Stat(self.CacheDir); os.IsNotExist(err) {
		if err := os.Mkdir(self.CacheDir, 0700); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile(self.CacheDir+"/"+path.Base(fileLocation), content, 0600); err != nil {
		return err
	}

	return nil

}

//func (self *repomd) Packages() []string {
// return
//}
