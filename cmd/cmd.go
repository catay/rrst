package cmd

import (
	"fmt"
	"github.com/catay/rrst/app"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

const (
	version       = "0.0.1"
	author        = "Steven Mertens <steven.mertens@catay.be>"
	defaultConfig = "/etc/rsst.yaml"
)

type cli struct {
	app        *kingpin.Application
	configFile *string
	listCmd    *kingpin.CmdClause
	syncCmd    *kingpin.CmdClause
	showCmd    *kingpin.CmdClause
}

func New() *cli {
	c := new(cli)
	c.app = kingpin.New("rsst", "Remote Repository Sync Tool")
	c.app.Version(version)
	c.app.Author(author)
	c.configFile = c.app.Flag("config", "Set path to configuration file.").Short('c').Default(defaultConfig).String()
	c.listCmd = c.app.Command("list", "List repository names and description.")
	c.syncCmd = c.app.Command("sync", "Synchronize remote to local repository sets.")
	c.showCmd = c.app.Command("show", "Show available repository sets.")

	return c
}

func (self *cli) Run() (err error) {
	args := kingpin.MustParse(self.app.Parse(os.Args[1:]))

	fmt.Println("config:", *self.configFile)

	r, err := app.New(*self.configFile)
	if err != nil {
		return err
	}

	switch args {
	case "sync":
		fmt.Println("command: sync")
		if err := r.Sync(); err != nil {
			return err
		}

	case "show":
		fmt.Println("command: show")
		r.Show()
	case "list":
		r.List()
	}

	return
}
