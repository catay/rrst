package rrst

import (
	"fmt"
	"github.com/catay/rrst/config"
)

const (
	cacheDir = "/var/cache/rrst"
)

type app struct {
	config *config.ConfigData
}

func New(configFile string) (a *app, err error) {
	a = new(app)
	a.config, err = config.New(configFile)
	if err != nil {
		return nil, err
	}

	// initialize default config variables
	a.setCacheDir()

	return
}

func (self *app) Print() {

	fmt.Println("cache_dir:", self.config.Globals.CacheDir)

	for _, r := range self.config.Repos {
		fmt.Println("*", r.Name)
	}
}

func (self *app) Sync() (err error) {
	fmt.Println("Not implemented yet!")
	return
}

func (self *app) setCacheDir() {
	if self.config.Globals.CacheDir == "" {
		self.config.Globals.CacheDir = cacheDir
	}
}
