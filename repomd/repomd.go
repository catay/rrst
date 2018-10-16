package repomd

import (
	"compress/gzip"
	"encoding/xml"
	"fmt"
	h "github.com/catay/rrst/util/http"
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
	LocalPath  string
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
func (r *Repomd) loadFromLocalRepoCache() error {
	data, err := ioutil.ReadFile(r.localRepoCacheFile)
	if err != nil {
		return err
	}

	err = xml.Unmarshal(data, r)
	return err
}

// Load the struct variables with data from the remote repomd XML file
// if present.
func (r *Repomd) loadFromRemoteRepoCache() ([]byte, error) {

	req, err := http.NewRequest("GET", r.remoteRepoCacheFile+"?"+r.Secret, nil)
	if err != nil {
		return nil, err
	}

	resp, err := h.HttpProxyGet(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := xml.Unmarshal(data, r); err != nil {
		return nil, err
	}

	return data, nil
}

// Create the cache directory if it doesn't exist.
// If it exists, do nothing.
func (r *Repomd) createCacheDir() error {
	_, err := os.Stat(r.CacheDir)
	if os.IsNotExist(err) {
		if err := os.Mkdir(r.CacheDir, 0700); err != nil {
			return err
		}
	}

	return nil
}

// Compare the revision of two repomd objets.
// Return true if the receiver has an outdated revision.
func (r *Repomd) hasOutdatedRevision(rm *Repomd) bool {
	if r.Revision < rm.Revision {
		return true
	}

	return false
}

// old methods

func (r *Repomd) Debug() {
	fmt.Printf("Url: %v\n", r.Url)
	fmt.Printf("Secret: %v\n", r.Secret)
}

func (r *Repomd) Metadata() error {

	ok, err := r.refreshRepomd()

	if err != nil {
		return err
	}

	if ok {
		for _, d := range r.Data {
			if err := r.fetchRepomdFile("/" + d.Location.Path); err != nil {
				return err
			}
		}

	}

	return r.unmarchalPrimaryData()
}

// Remove the cache directory and content for the repository.
func (r *Repomd) Clean() error {
	return os.RemoveAll(r.CacheDir)
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

func (r *Repomd) refreshRepomd() (bool, error) {

	ok := false

	t := &Repomd{
		Url:                 r.Url,
		Secret:              r.Secret,
		CacheDir:            r.CacheDir,
		localRepoCacheFile:  r.CacheDir + "/repomd.xml",
		remoteRepoCacheFile: r.Url + "/repodata/repomd.xml",
	}

	content, err := t.loadFromRemoteRepoCache()
	if err != nil {
		return ok, err
	}

	// load local cache if present
	if err := r.loadFromLocalRepoCache(); err != nil {
		if !os.IsNotExist(err) {
			return ok, err
		}
	}

	if r.hasOutdatedRevision(t) {
		ok = true
		fmt.Println("Refresh repomd cache ...")
		// clean current cache if present
		if err := r.Clean(); err != nil {
			if !os.IsNotExist(err) {
				return ok, err
			}
		}

		*r = *t

		if err := r.createCacheDir(); err != nil {
			return ok, err
		}

		if err := ioutil.WriteFile(r.localRepoCacheFile, content, 0600); err != nil {
			return ok, err
		}
	}

	return ok, nil
}

func (r *Repomd) fetchRepomdFile(fileLocation string) error {

	req, err := http.NewRequest("GET", r.Url+fileLocation+"?"+r.Secret, nil)
	if err != nil {
		return err
	}

	resp, err := h.HttpProxyGet(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// check if dir exists, if not create it
	if _, err := os.Stat(r.CacheDir); os.IsNotExist(err) {
		if err := os.Mkdir(r.CacheDir, 0700); err != nil {
			return err
		}
	}

	return ioutil.WriteFile(r.CacheDir+"/"+path.Base(fileLocation), content, 0600)

}

func (r *Repomd) removeRepomdFile(fileLocation string) error {
	return os.Remove(r.CacheDir + "/" + path.Base(fileLocation))
}

func (r *Repomd) unmarchalPrimaryData() error {
	var primaryCache string

	for _, d := range r.Data {
		if d.Type == "primary" {
			primaryCache = r.CacheDir + "/" + path.Base(d.Location.Path)
		}
	}

	f, err := os.Open(primaryCache)
	if err != nil {
		return err
	}

	pc, err := gzip.NewReader(f)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(pc)
	if err != nil {
		return err
	}

	return xml.Unmarshal(data, &r.PrimaryData)
}

func (r *Repomd) Packages() []RpmPackage {
	return r.PrimaryData.Package
}

func (r *Repomd) PackageCount() string {
	return r.PrimaryData.Packages
}
