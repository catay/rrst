package app

import (
	"fmt"
	"github.com/catay/rrst/config"
)

const (
	cacheDir = "/var/cache/rrst"
)

type App struct {
	config *config.ConfigData
}

func New(configFile string) (a *App, err error) {
	a = new(App)
	a.config, err = config.NewConfig(configFile)
	if err != nil {
		return nil, err
	}

	// initialize default config variables
	a.setCacheDir()

	return
}

func (self *App) Print() {

	fmt.Println("cache_dir:", self.config.Globals.CacheDir)

	for _, r := range self.config.Repos {
		fmt.Println("*", r.Name)
		fmt.Println(" -", r.CacheDir)
	}
}

func (self *App) Sync() (err error) {
	fmt.Println("Not implemented yet!")
	return
}

func (self *App) setCacheDir() {
	if self.config.Globals.CacheDir == "" {
		self.config.Globals.CacheDir = cacheDir
	}
}
