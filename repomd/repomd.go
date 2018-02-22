package repomd

import (
	"compress/gzip"
	"encoding/xml"
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
	Revision    string       `xml:"revision"`
	Data        []repomdData `xml:"data"`
	PrimaryData primaryData
}

type repomdData struct {
	Type     string             `xml:"type,attr"`
	Size     string             `xml:"size"`
	Location repomdDataLocation `xml:"location"`
}

type repomdDataLocation struct {
	Path string `xml:"href,attr"`
}

type primaryData struct {
	Packages string       `xml:"packages,attr"`
	Package  []RpmPackage `xml:"package"`
}

type RpmPackage struct {
	Type       string      `xml:"type,attr"`
	Name       string      `xml:"name"`
	Arch       string      `xml:"arch"`
	Loc        rpmLocation `xml:"location"`
	ToDownload bool
}

type rpmLocation struct {
	Path string `xml:"href,attr"`
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

	ok, err := self.refreshRepomd()

	if err != nil {
		return err
	}

	if ok {
		for _, d := range self.Data {
			fmt.Printf("DEBUG - location: %v\n", d.Location.Path)
			if err := self.fetchRepomdFile("/" + d.Location.Path); err != nil {
				return err
			}
		}

	}

	if err := self.unmarchalPrimaryData(); err != nil {
		return err
	}

	return nil
}

func (self *repomd) unMarshalRepomdData() error {

	data, err := ioutil.ReadFile(self.CacheDir + "/repomd.xml")
	if err != nil {
		return err
	}

	if err := xml.Unmarshal(data, self); err != nil {
		return err
	}

	return nil
}

// 0.  fetch remote repodata.xml in memory
// 1. check if local repodata.xml exists
// if exists
//    unmarchal it in property value
//    unmarchal in temp variable
//    compare old revision with new revision
//    if newer revision, store to disk and fetch other data files
// if not exists
//    store to disk and fetch other data files

func (self *repomd) refreshRepomd() (bool, error) {

	repomdFile := self.CacheDir + "/repomd.xml"
	ok := false

	resp, err := http.Get(self.Url + "/repodata/repomd.xml" + "?" + self.Secret)
	if err != nil {
		return ok, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ok, errors.New(fmt.Sprintf("HTTP error %v ", resp.StatusCode))
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ok, err
	}

	t := *self

	if err := xml.Unmarshal(content, &t); err != nil {
		return ok, err
	}

	f, err := ioutil.ReadFile(repomdFile)

	if err == nil {
		if err := xml.Unmarshal(f, self); err != nil {
			return ok, err
		}
	} else {
		if !os.IsNotExist(err) {
			return ok, err
		}
	}

	if t.Revision != self.Revision {
		ok = true
		fmt.Println("DEBUG: repomd REVISION OUTDATED !!!")
		*self = t

		// check if dir exists, if not create it
		if _, err := os.Stat(self.CacheDir); os.IsNotExist(err) {
			if err := os.Mkdir(self.CacheDir, 0700); err != nil {
				return ok, err
			}
		}

		if err := ioutil.WriteFile(repomdFile, content, 0600); err != nil {
			return ok, err
		}
	}

	return ok, nil
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

func (self *repomd) unmarchalPrimaryData() error {
	var primaryCache string

	for _, d := range self.Data {
		if d.Type == "primary" {
			primaryCache = self.CacheDir + "/" + path.Base(d.Location.Path)
		}
	}

	f, err := os.Open(primaryCache)
	if err != nil {
		return err
	}

	r, err := gzip.NewReader(f)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	if err := xml.Unmarshal(data, &self.PrimaryData); err != nil {
		return err
	}

	return nil
}

func (self *repomd) Packages() []RpmPackage {
	return self.PrimaryData.Package
}

func (self *repomd) PackageCount() string {
	return self.PrimaryData.Packages
}
