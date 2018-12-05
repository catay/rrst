package app

import (
	"fmt"
	"github.com/catay/rrst/config"
	"github.com/catay/rrst/repository"
	"os"
)

type App struct {
	config       *config.Config
	repositories []*repository.Repository
}

func NewApp(configFile string) (a *App, err error) {
	a = &App{}

	// initialize the configuration
	a.config, err = config.NewConfig(configFile)
	if err != nil {
		return nil, err
	}

	// initialize the root content directory
	if err := a.initContentPath(); err != nil {
		return nil, fmt.Errorf("init content path failed: %s", err)
	}

	// initialize repositories
	for i, _ := range a.config.RepoConfigs {
		a.repositories = append(a.repositories, repository.NewRepository(a.config.RepoConfigs[i]))
	}

	return a, nil
}

func (a *App) Create(action string) {
	fmt.Println(action)
}

func (a *App) List(repo string) {

	if len(a.repositories) == 0 {
		fmt.Println("No repositories configured.")
		return
	}

	if repo != "" {
		if r, ok := a.getRepoName(repo); ok {
			fmt.Println(r.Id, r.Name, r.ContentSuffixPath, len(r.SystemTags))
		} else {
			fmt.Println("No configured repository", repo, "found.")
		}
	} else {
		for _, r := range a.repositories {
			fmt.Println(r.Id, r.Name, r.ContentSuffixPath, len(r.SystemTags))
		}
	}
}

func (a *App) Update(repo string) {
	if len(a.repositories) == 0 {
		fmt.Println("No repositories configured.")
		return
	}

	if repo != "" {
		if r, ok := a.getRepoName(repo); ok {
			fmt.Printf("* Updating %s ...\n", r.Name)
			r.Update()
		} else {
			fmt.Println("No configured repository", repo, "found.")
		}
	} else {
		for _, r := range a.repositories {
			fmt.Printf("* Updating %s ...\n", r.Name)
			r.Update()
		}
	}
}

func (a *App) Delete(action string) {
	fmt.Println(action)
}

func (a *App) isRepoName(repo string) bool {
	for _, r := range a.repositories {
		if repo == r.Name {
			return true
		}
	}
	return false
}

func (a *App) getRepoName(repo string) (*repository.Repository, bool) {
	for i, r := range a.repositories {
		if repo == r.Name {
			return a.repositories[i], true
		}
	}
	return nil, false
}

func (a *App) initContentPath() error {
	if err := os.MkdirAll(a.config.GlobalConfig.ContentPath, 0700); err != nil {
		return err
	}
	return nil
}
