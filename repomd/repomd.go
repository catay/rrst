package repomd

import (
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
	Revision string       `xml:"revision"`
	Data     []repomdData `xml:"data"`
}

type repomdData struct {
	Type     string             `xml:"type,attr"`
	Size     string             `xml:"size"`
	Location repomdDataLocation `xml:"location"`
}

type repomdDataLocation struct {
	HRef string `xml:"href,attr"`
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
			fmt.Printf("DEBUG - location: %v\n", d.Location.HRef)
			if err := self.fetchRepomdFile("/" + d.Location.HRef); err != nil {
				return err
			}
		}

	}

	//	if err := self.fetchRepomdFile("/repodata/repomd.xml"); err != nil {
	//		return err
	//	}

	//	if err := self.unMarshalRepomdData(); err != nil {
	//		return err
	//	}
	//

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

	// 0.  fetch remote repodata.xml in memory
	// 1. check if local repodata.xml exists
	// if exists
	//    unmarchal it in property value
	//    unmarchal in temp variable
	//    compare old revision with new revision
	//    if newer revision, store to disk and fetch other data files
	// if not exists
	//    store to disk and fetch other data files

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
