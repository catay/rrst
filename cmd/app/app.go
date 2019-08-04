package app

import (
	"fmt"
	"github.com/catay/rrst/config"
	"github.com/catay/rrst/repository"
	"github.com/catay/rrst/server"
	"os"
	"strings"
	"text/tabwriter"
)

const (
	DefaultConfig = config.DefaultConfigPath
	DefaultPort   = config.DefaultServerPort
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
		r, err := repository.NewRepository(a.config.RepoConfigs[i])
		if err != nil {
			return nil, fmt.Errorf("Initialize repo %s failed: %s", a.config.RepoConfigs[i].Name, err)
		}
		a.repositories = append(a.repositories, r)
	}

	return a, nil
}

func (a *App) Create(action string) {
	fmt.Println(action)
}

func (a *App) Status(repo string) {
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

func (a *App) List(repo string, tagsOrRevs ...string) {
	if len(a.repositories) == 0 {
		fmt.Println("No repositories configured.")
		return
	}

	if r, ok := a.getRepoName(repo); ok {
		// if only no tag or revision is provided, compare with latest tag
		if len(tagsOrRevs) == 0 {
			tagsOrRevs = append(tagsOrRevs, "latest")
		}
		packageMap, err := r.PackageVersions(tagsOrRevs...)
		if err != nil {
			fmt.Println("list error: ", err)
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
		fmt.Fprintf(w, "PACKAGE\t%v\n", strings.Join(tagsOrRevs, "\t"))
		for k, v := range packageMap {
			fmt.Fprintf(w, "%v\t%v\n", k, strings.Join(v, "\t"))
		}
		w.Flush()
	} else {
		fmt.Println("No configured repository", repo, "found.")
	}
}

func (a *App) Update(repo string, rev int64) {
	if len(a.repositories) == 0 {
		fmt.Println("No repositories configured.")
		return
	}

	if repo != "" {
		if r, ok := a.getRepoName(repo); ok {
			//fmt.Printf("* Updating %s ...\n", r.Name)
			_, err := r.Update(rev)
			if err != nil {
				fmt.Println(" > error: ", err)
			}
		} else {
			fmt.Println("No configured repository", repo, "found.")
		}
	} else {
		for _, r := range a.repositories {
			//fmt.Printf("* Updating %s ...\n", r.Name)
			r.Update(rev) // rev will always be empty
		}
	}
}

func (a *App) Tag(repo string, tag string, rev int64, force bool) {
	if len(a.repositories) == 0 {
		fmt.Println("No repositories configured.")
		return
	}

	if repo != "" {
		if r, ok := a.getRepoName(repo); ok {
			_, err := r.Tag(tag, rev, force)
			if err != nil {
				fmt.Println("tag error: ", err)
			}
		} else {
			fmt.Println("No configured repository", repo, "found.")
		}
	}
}

func (a *App) Diff(repo string, tags ...string) {
	if len(a.repositories) == 0 {
		fmt.Println("No repositories configured.")
		return
	}

	if repo != "" {
		if r, ok := a.getRepoName(repo); ok {
			// if only 1 tag is provided, compare with latest tag
			if len(tags) == 1 {
				tags = append(tags, "latest")
			}
			diff, err := r.Diff(tags...)
			if err != nil {
				fmt.Println("diff error: ", err)
				return
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
			fmt.Fprintf(w, "PACKAGE\t%v\n", strings.Join(tags, "\t"))
			for k, v := range diff {
				// only show packages with a different version in a tagged revision
				var show bool
				for _, a := range v[1:] {
					if v[0] != a {
						show = true
						break
					}
				}

				if show {
					// set the version string to - when the package is not present in a
					// tagged revision
					for i := range v {
						if v[i] == "" {
							v[i] = "-"
						}
					}

					fmt.Fprintf(w, "%v\t%v\n", k, strings.Join(v, "\t"))
				}
			}
			w.Flush()
		}

	} else {
		fmt.Println("No configured repository", repo, "found.")
	}
}

func (a *App) Delete(action string) {
	fmt.Println(action)
}

func (a *App) Server(port string) error {
	s := server.NewServer(port, a.repositories)
	return s.Run()
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
			fmt.Fprintln(w, "REVISION\tCREATED\tTAGS")
			for _, v := range r.Revisions {
				tags := strings.Join(v.TagNames(), ", ")
				if tags == "" {
					tags = "<none>"
				}
				fmt.Fprintf(w, "%v\t%v\t%v\n", v.Id, v.Timestamp(), tags)
			}
		} else {
			fmt.Fprintf(w, "No revisions available for repository %v\n.", repo)
		}
		return w.Flush()

	} else {
		fmt.Printf("Repository '%v' not found.\n", repo)
	}
	return nil
}

// The showRepos method prints general repository information to
// standard ouptput of all the configured repositories.
func (a *App) showRepos() error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 4, ' ', 0)
	fmt.Fprintln(w, "ID\tREPOSITORY\tENABLED\t#REVISIONS\t#TAGS\tUPDATED")
	for _, r := range a.repositories {
		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\n", r.Id, r.Name, r.Enabled, len(r.Revisions), len(r.Tags), r.LastUpdated())
	}
	return w.Flush()
}
