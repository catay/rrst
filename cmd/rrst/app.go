package app

import (
	"fmt"
	"github.com/catay/rrst/config"
	"github.com/catay/rrst/repository"
	"os"
	"text/tabwriter"
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
		a.showRepo(repo)
	} else {
		a.showRepos()
	}
}

func (a *App) Update(repo string, rev string) {
	if len(a.repositories) == 0 {
		fmt.Println("No repositories configured.")
		return
	}

	if repo != "" {
		if r, ok := a.getRepoName(repo); ok {
			fmt.Printf("* Updating %s ...\n", r.Name)
			_, err := r.Update(rev)
			if err != nil {
				fmt.Println(" > error: ", err)
			}
		} else {
			fmt.Println("No configured repository", repo, "found.")
		}
	} else {
		for _, r := range a.repositories {
			fmt.Printf("* Updating %s ...\n", r.Name)
			r.Update(rev) // rev will always be empty
		}
	}
}

func (a *App) Tag(repo string, tag string, rev string, force bool) {
	if len(a.repositories) == 0 {
		fmt.Println("No repositories configured.")
		return
	}

	if repo != "" {
		if r, ok := a.getRepoName(repo); ok {
			fmt.Printf("* Tag %s ...\n", r.Name)
			_, err := r.Tag(tag, rev, force)
			if err != nil {
				fmt.Println("tag error: ", err)
			}
		} else {
			fmt.Println("No configured repository", repo, "found.")
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

// The showRepo method prints detailed repository information to
// standard ouptput of the specified repository when present.
func (a *App) showRepo(repo string) error {
	if r, ok := a.getRepoName(repo); ok {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
		if r.HasRevisions() {
			fmt.Fprintln(w, "REVISIONS\tCREATED\tTAGS")
			for _, v := range r.Revisions {
				fmt.Fprintf(w, "%v\t%v\t%v\n", v, v.Timestamp(), r.RevisionTags(v))
			}
		} else {
			fmt.Fprintln(w, "No revisions available for repository.", repo)
		}
		return w.Flush()

	} else {
		fmt.Println("No configured repository", repo, "found.")
	}
	return nil
}

// The showRepos method prints general repository information to
// standard ouptput of all the configured repositories.
func (a *App) showRepos() error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Fprintln(w, "ID\tREPOSITORY\tENABLED\t#REVISIONS\t#TAGS\tUPDATED")
	for _, r := range a.repositories {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\n", r.Id, r.Name, r.Enabled, len(r.Revisions), len(r.Tags()), r.LastUpdated())
	}
	return w.Flush()
}
