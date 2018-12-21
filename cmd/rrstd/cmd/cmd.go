package cmd

import (
	"github.com/catay/rrst/cmd/rrstd/app"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"path/filepath"
)

// the version string will be injected during the build.
var Version = "0.0.0"

const (
	author = "Steven Mertens <steven.mertens@catay.be>"
)

type Cli struct {
	*kingpin.Application
	app        *app.App
	configFile *string
	verbose    *bool
}

func NewCli() *Cli {
	c := &Cli{
		Application: kingpin.New(filepath.Base(os.Args[0]), "Remote Repository Sync Tool Daemon"),
	}
	c.Version(Version)
	c.Author(author)
	c.configFile = c.Flag("config", "Path to alternate YAML configuration file.").Short('c').Default(app.DefaultConfig).String()
	c.verbose = c.Flag("verbose", "Turn on verbose output. Default is verbose turned off.").Short('v').Bool()
	return c
}

func (c *Cli) Run() error {
	var err error
	//c.action = kingpin.MustParse(c.Parse(os.Args[1:]))
	kingpin.MustParse(c.Parse(os.Args[1:]))
	c.app, err = app.NewApp(*c.configFile)
	if err != nil {
		return err
	}

	return c.app.Server()
}
