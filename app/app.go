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

func NewApp(configFile string) (a *App, err error) {
	a = new(App)
	a.config, err = config.NewConfig(configFile)
	if err != nil {
		return nil, err
	}

	// initialize default config variables
	a.setCacheDir()

	return
}

func (self *App) List() {

	fmt.Println("Configured repositories:")

	if len(self.config.Repos) > 0 {
		for _, r := range self.config.Repos {
			fmt.Println("  ", r.Name)
		}
	} else {
		fmt.Println("  No configured repositories found.")
	}
}

func (self *App) Show() {

	fmt.Println("cache_dir:", self.config.Globals.CacheDir)

	for _, r := range self.config.Repos {
		fmt.Println("*", r.Name)
		fmt.Println(" -", r.CacheDir)
	}
}

func (self *App) Sync() (err error) {
	for _, r := range self.config.Repos {
		fmt.Println("*", r.Name)
		if err := r.Sync(); err != nil {
			return err
		}
	}
	return nil
}

func (self *App) Clean(repoName string) {

	fmt.Println("Clean cache repositories:")

	var found bool

	if repoName != "" {
		for i := range self.config.Repos {
			if self.config.Repos[i].Name == repoName {
				fmt.Println("Clean", repoName)
				self.config.Repos[i].Clean()
				found = true
				break

			}
		}

		if !found {
			fmt.Println("  No configured repository", repoName, "found.")
		}
	} else {
		if len(self.config.Repos) > 0 {
			fmt.Println("Clean all")
			for i := range self.config.Repos {
				self.config.Repos[i].Clean()
			}

		} else {
			fmt.Println("  No configured repositories found.")
		}
	}
}

func (self *App) setCacheDir() {
	if self.config.Globals.CacheDir == "" {
		self.config.Globals.CacheDir = cacheDir
	}
}
