package suse

import (
	"encoding/json"
	"github.com/catay/rrst/util"
	h "github.com/catay/rrst/util/http"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type SCCApi struct {
	apiURI               string
	regCode              string
	cacheDir             string
	cacheFile            string
	cacheRefreshDuration int
}

type Product struct {
	Name         string       `json:"name"`
	FriendlyName string       `json:"friendly_name"`
	Repos        []Repository `json:"repositories"`
	Extensions   []Product    `json:"extensions"`
}

type Repository struct {
	Name    string `json:"name"`
	Url     string `json:"url"`
	Enabled bool   `json:"enabled"`
}

func NewSCCApi(regCode string, cacheDir string) (self *SCCApi) {
	api := &SCCApi{}
	api.apiURI = "https://scc.suse.com/connect/subscriptions/products.json"
	api.regCode = regCode
	api.cacheDir = cacheDir
	api.cacheFile = api.cacheDir + "/" + util.Sha256Sum(api.regCode)
	api.cacheRefreshDuration = 86400
	return api
}

// public

//func (self *SCCApi) FetchProductsJson(force bool) {
func (self *SCCApi) FetchProductsJson() error {

	if !self.isCacheExpired() {
		return nil
	}

	req, err := http.NewRequest("GET", self.apiURI, nil)
	if err != nil {
		return err
	}

	// set SCC registration code as a header for authentication
	req.Header.Add("Authorization", "Token token="+self.regCode)

	resp, err := h.HttpProxyGet(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(self.cacheFile, body, 0600); err != nil {
		return err
	}

	return nil
}

func (self *SCCApi) GetSecretURI(url string) (string, bool) {
	return self.getRepoSecret(self.getProducts(), url)
}

// private

func (self *SCCApi) isCacheExpired() bool {
	f, err := os.Open(self.cacheFile)
	defer f.Close()

	if err != nil {
		return true
	} else {
		fi, err := f.Stat()
		if err != nil {
			return true
		}

		if time.Since(fi.ModTime()).Seconds() > float64(self.cacheRefreshDuration) {
			return true
		}
	}

	return false
}

func (self *SCCApi) getProducts() []Product {
	var p []Product

	data, err := ioutil.ReadFile(self.cacheFile)
	if err != nil {
		log.Fatal(err)
	}

	if err := json.Unmarshal(data, &p); err != nil {
		log.Fatal(err)
	}

	return p
}

func (self *SCCApi) getRepoSecret(p []Product, url string) (string, bool) {

	var secret string
	var ok bool

	for _, v := range p {
		for _, r := range v.Repos {
			u := strings.SplitN(r.Url, "/?", 2)[0]
			if r.Enabled && u == url {
				secret = strings.SplitN(r.Url, "/?", 2)[1]
				ok = true
			}
		}
		if !ok {
			secret, ok = self.getRepoSecret(v.Extensions, url)
		}
	}

	return secret, ok
}
