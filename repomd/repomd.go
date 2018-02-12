package repomd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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
	resp, err := http.Get(self.Url + "/repodata/repomd.xml" + "?" + self.Secret)
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

	if err := os.Mkdir(self.CacheDir, 0700); err != nil {
		return err
	}

	if err := ioutil.WriteFile(self.CacheDir+"/repomd.xml", content, 0600); err != nil {
		return err
	}

	return nil
}

//func (self *repomd) Packages() []string {
// return
//}
