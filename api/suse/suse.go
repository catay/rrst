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

func NewSCCApi(regCode string, cacheDir string) (s *SCCApi) {
	s = &SCCApi{}
	s.apiURI = "https://scc.suse.com/connect/subscriptions/products.json"
	s.regCode = regCode
	s.cacheDir = cacheDir
	s.cacheFile = s.cacheDir + "/" + util.Sha256Sum(s.regCode)
	s.cacheRefreshDuration = 86400
	return s
}

// public

//func (s *SCCApi) FetchProductsJson(force bool) {
func (s *SCCApi) FetchProductsJson() error {

	if !s.isCacheExpired() {
		return nil
	}

	req, err := http.NewRequest("GET", s.apiURI, nil)
	if err != nil {
		return err
	}

	// set SCC registration code as a header for authentication
	req.Header.Add("Authorization", "Token token="+s.regCode)

	resp, err := h.HttpProxyGet(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(s.cacheFile, body, 0600)
	return err
}

func (s *SCCApi) GetSecretURI(url string) (string, bool) {
	return s.getRepoSecret(s.getProducts(), url)
}

// private

func (s *SCCApi) isCacheExpired() bool {
	f, err := os.Open(s.cacheFile)
	defer f.Close()

	if err != nil {
		return true
	} else {
		fi, err := f.Stat()
		if err != nil {
			return true
		}

		if time.Since(fi.ModTime()).Seconds() > float64(s.cacheRefreshDuration) {
			return true
		}
	}

	return false
}

func (s *SCCApi) getProducts() []Product {
	var p []Product

	data, err := ioutil.ReadFile(s.cacheFile)
	if err != nil {
		log.Fatal(err)
	}

	if err := json.Unmarshal(data, &p); err != nil {
		log.Fatal(err)
	}

	return p
}

func (s *SCCApi) getRepoSecret(p []Product, url string) (string, bool) {

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
			secret, ok = s.getRepoSecret(v.Extensions, url)
		}
	}

	return secret, ok
}
