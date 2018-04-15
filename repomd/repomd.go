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

type Repomd struct {
	Url                 string
	Secret              string
	CacheDir            string
	localRepoCacheFile  string
	remoteRepoCacheFile string
	Revision            string       `xml:"revision"`
	Data                []repomdData `xml:"data"`
	PrimaryData         primaryData
}

type repomdData struct {
	Type      string             `xml:"type,attr"`
	Size      string             `xml:"size"`
	Timestamp string             `xml:"timestamp"`
	Location  repomdDataLocation `xml:"location"`
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

// public methods

// repomd constructor
func NewRepoMd(url, secret string, cacheDir string) (*Repomd, error) {
	r := &Repomd{
		Url:                 url,
		Secret:              secret,
		CacheDir:            cacheDir,
		localRepoCacheFile:  cacheDir + "/repomd.xml",
		remoteRepoCacheFile: url + "/repodata/repomd.xml",
	}

	// load local cache if present
	if err := r.loadFromLocalRepoCache(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return r, nil
}

// private methods

// Load the struct variables with data from the locally cached
// repomd XML file if present.
func (self *Repomd) loadFromLocalRepoCache() error {
	data, err := ioutil.ReadFile(self.localRepoCacheFile)
	if err != nil {
		return err
	}

	if err := xml.Unmarshal(data, self); err != nil {
		return err
	}

	return nil
}

// Load the struct variables with data from the remote repomd XML file
// if present.
func (self *Repomd) loadFromRemoteRepoCache() ([]byte, error) {

	resp, err := http.Get(self.remoteRepoCacheFile + "?" + self.Secret)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("HTTP error %v ", resp.StatusCode))
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := xml.Unmarshal(data, self); err != nil {
		return nil, err
	}

	return data, nil
}

// Create the cache directory if it doesn't exist.
// If it exists, do nothing.
func (self *Repomd) createCacheDir() error {
	_, err := os.Stat(self.CacheDir)
	if os.IsNotExist(err) {
		if err := os.Mkdir(self.CacheDir, 0700); err != nil {
			return err
		}
	}

	return nil
}

// Compare the revision of two repomd objets.
// Return true if the receiver has an outdated revision.
func (self *Repomd) hasOutdatedRevision(r *Repomd) bool {
	if self.Revision < r.Revision {
		return true
	}

	return false
}

// old methods

func (self *Repomd) Debug() {
	fmt.Printf("Url: %v\n", self.Url)
	fmt.Printf("Secret: %v\n", self.Secret)
}

func (self *Repomd) Metadata() error {

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

func (self *Repomd) Clean() error {

	if err := os.Remove(self.localRepoCacheFile); err != nil {
		return err
	}

	for _, d := range self.Data {
		fmt.Printf("DEBUG clean - location: %v\n", d.Location.Path)
		if err := self.removeRepomdFile("/" + d.Location.Path); err != nil {
			return err
		}
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

func (self *Repomd) refreshRepomd() (bool, error) {

	ok := false

	t := &Repomd{
		Url:                 self.Url,
		Secret:              self.Secret,
		CacheDir:            self.CacheDir,
		localRepoCacheFile:  self.CacheDir + "/repomd.xml",
		remoteRepoCacheFile: self.Url + "/repodata/repomd.xml",
	}

	content, err := t.loadFromRemoteRepoCache()
	if err != nil {
		return ok, err
	}

	// load local cache if present
	if err := self.loadFromLocalRepoCache(); err != nil {
		if !os.IsNotExist(err) {
			return ok, err
		}
	}

	if self.hasOutdatedRevision(t) {
		ok = true
		fmt.Println("DEBUG: repomd REVISION OUTDATED !!!")
		// clean current cache if present
		if err := self.Clean(); err != nil {
			if !os.IsNotExist(err) {
				return ok, err
			}
		}

		*self = *t

		if err := self.createCacheDir(); err != nil {
			return ok, err
		}

		if err := ioutil.WriteFile(self.localRepoCacheFile, content, 0600); err != nil {
			return ok, err
		}
	}

	return ok, nil
}

func (self *Repomd) fetchRepomdFile(fileLocation string) error {

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

func (self *Repomd) removeRepomdFile(fileLocation string) error {
	if err := os.Remove(self.CacheDir + "/" + path.Base(fileLocation)); err != nil {
		return err
	}

	return nil
}

func (self *Repomd) unmarchalPrimaryData() error {
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

func (self *Repomd) Packages() []RpmPackage {
	return self.PrimaryData.Package
}

func (self *Repomd) PackageCount() string {
	return self.PrimaryData.Packages
}
