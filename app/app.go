package app

import (
	"fmt"
	"github.com/catay/rrst/config"
	"os"
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

	// set proxy
	a.setProxy()

	// initialize default config variables
	a.setCacheDir()
	if err := a.mkCacheDir(); err != nil {
		return nil, err
	}

	return
}

func (a *App) List() {

	fmt.Println("Configured repositories:")

	if len(a.config.Repos) > 0 {
		for _, r := range a.config.Repos {
			fmt.Println("  ", r.Name)
		}
	} else {
		fmt.Println("  No configured repositories found.")
	}
}

func (a *App) Show() {

	fmt.Println("cache_dir:", a.config.Globals.CacheDir)

	for _, r := range a.config.Repos {
		fmt.Println("*", r.Name)
		fmt.Println(" -", r.CacheDir)
	}
}

func (a *App) Sync(repoName string) (err error) {

	fmt.Println("Sync repositories:")

	if repoName != "" {
		if r := a.config.GetRepoByName(repoName); r != nil {
			if err := r.Sync(); err != nil {
				return err
			}
		} else {
			fmt.Println("  No configured repository", repoName, "found.")
		}
	} else {
		if len(a.config.Repos) > 0 {
			for i := range a.config.Repos {
				if err := a.config.Repos[i].Sync(); err != nil {
					fmt.Println("    ", err)
				}
			}

		} else {
			fmt.Println("  No configured repositories found.")
		}
	}

	return nil
}

func (a *App) Clean(repoName string) (err error) {

	fmt.Println("Clean cache repositories:")

	if repoName != "" {
		if r := a.config.GetRepoByName(repoName); r != nil {
			if err := r.Clean(); err != nil {
				return err
			}
		} else {
			fmt.Println("  No configured repository", repoName, "found.")
		}
	} else {
		if len(a.config.Repos) > 0 {
			for i := range a.config.Repos {
				if err := a.config.Repos[i].Clean(); err != nil {
					return err
				}
			}

		} else {
			fmt.Println("  No configured repositories found.")
		}
	}

	return nil
}

func (a *App) setCacheDir() {
	if a.config.Globals.CacheDir == "" {
		a.config.Globals.CacheDir = cacheDir
	}
}

func (a *App) mkCacheDir() (err error) {
	return os.MkdirAll(a.config.Globals.CacheDir, 0755)
}

func (a *App) setProxy() (err error) {
	if a.config.Globals.ProxyURL != "" {
		if err := os.Setenv("http_proxy", a.config.Globals.ProxyURL); err != nil {
			return err
		}
	}
	return nil
}
