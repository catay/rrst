package cmd

import (
	"github.com/catay/rrst/cmd/rrstd/app"
	"github.com/catay/rrst/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"path/filepath"
)

type Cli struct {
	*kingpin.Application
	app        *app.App
	configFile *string
	verbose    *bool
	port       *string
}

func NewCli() *Cli {
	c := &Cli{
		Application: kingpin.New(filepath.Base(os.Args[0]), "Remote Repository Sync Tool Daemon"),
	}
	c.Version(version.Version)
	c.Author(version.Author)
	c.configFile = c.Flag("config", "Path to alternate YAML configuration file.").Short('c').Default(app.DefaultConfig).String()
	c.verbose = c.Flag("verbose", "Turn on verbose output. Default is verbose turned off.").Short('v').Bool()
	c.port = c.Flag("port", "Port number to listen on.").Short('p').Default(app.DefaultPort).String()
	return c
}

func (c *Cli) Run() error {
	var err error
	//c.action = kingpin.MustParse(c.Parse(os.Args[1:]))
	kingpin.MustParse(c.Parse(os.Args[1:]))
	c.app, err = app.NewApp(*c.configFile, *c.port)
	if err != nil {
		return err
	}

	return c.app.Server()
}
