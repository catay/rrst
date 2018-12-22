package app

import (
	"fmt"
	"github.com/catay/rrst/config"
	"github.com/catay/rrst/repository"
	"github.com/gorilla/handlers"
	"net/http"
	"os"
)

const (
	DefaultConfig = config.DefaultConfigPath
	DefaultPort   = config.DefaultServerPort
)

type App struct {
	config       *config.Config
	repositories []*repository.Repository
	port         string
}

func NewApp(configFile string, port string) (a *App, err error) {
	a = &App{port: port}

	// initialize the configuration
	a.config, err = config.NewConfig(configFile)
	if err != nil {
		return nil, err
	}

	// initialize repositories
	for i, _ := range a.config.RepoConfigs {
		a.repositories = append(a.repositories, repository.NewRepository(a.config.RepoConfigs[i]))
	}

	return a, nil
}

func (a *App) Server() error {
	fmt.Println("Start server")
	for _, r := range a.repositories {

		if len(r.Tags) != 0 && r.Enabled {
			fmt.Println("* start handler for repo", r.Name)
			for _, t := range r.Tags {
				fmt.Println("  > tag: ", t.Name)
				// serve metadata
				serveMdPath := "/" + r.ContentSuffixPath + "/" + t.Name + "/repodata/"
				localMdPath := r.ContentTagsPath + "/" + t.Name + "/repodata/"

				http.Handle(serveMdPath, handlers.CombinedLoggingHandler(
					os.Stdout,
					http.StripPrefix(serveMdPath,
						http.FileServer(http.Dir(localMdPath))),
				))

				// serve files
				serveFilesPath := "/" + r.ContentSuffixPath + "/" + t.Name + "/"
				localFilesPath := r.ContentFilesPath + "/"
				http.Handle(serveFilesPath, handlers.CombinedLoggingHandler(
					os.Stdout, http.StripPrefix(serveFilesPath,
						http.FileServer(http.Dir(localFilesPath))),
				))
			}
		}
	}

	return http.ListenAndServe(":"+a.port, nil)
}
