package cmd

import (
	//"fmt"
	"github.com/catay/rrst/cmd/rrst"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"path/filepath"
)

// the version string will be injected during the build.
var Version = "0.0.0"

const (
	author        = "Steven Mertens <steven.mertens@catay.be>"
	defaultConfig = "/etc/rsst.yaml"
)

type Cli struct {
	*kingpin.Application
	app              *app.App
	action           string
	configFile       *string
	verbose          *bool
	cmdCreate        *kingpin.CmdClause
	cmdList          *kingpin.CmdClause
	cmdUpdate        *kingpin.CmdClause
	cmdTag           *kingpin.CmdClause
	cmdDelete        *kingpin.CmdClause
	cmdTagForceFlag  *bool
	cmdCreateRepoArg *string
	cmdListRepoArg   *string
	cmdUpdateRepoArg *string
	cmdUpdateRevArg  *int64
	cmdTagRepoArg    *string
	cmdTagTagArg     *string
	cmdTagRevArg     *int64
	cmdDeleteRepoArg *string
}

func NewCli() *Cli {
	c := &Cli{
		Application: kingpin.New(filepath.Base(os.Args[0]), "Remote Repository Sync Tool"),
	}
	c.Version(Version)
	c.Author(author)
	c.configFile = c.Flag("config", "Path to alternate YAML configuration file.").Short('c').Default(defaultConfig).String()
	c.verbose = c.Flag("verbose", "Turn on verbose output. Default is verbose turned off.").Short('v').Bool()
	c.cmdCreate = c.Command("create", "Create custom repositories. **NOT IMPLEMENTED**")
	c.cmdList = c.Command("list", "List repositories, revisions and tags.")
	c.cmdUpdate = c.Command("update", "Update repositories with upstream content.")
	c.cmdTag = c.Command("tag", "Tag repository revisions.")
	c.cmdDelete = c.Command("delete", "Delete repositories. **NOT IMPLEMENTED**")

	c.cmdCreateRepoArg = c.cmdCreate.Arg("repo name", "Repository name.").String()
	c.cmdListRepoArg = c.cmdList.Arg("repo name", "Repository name.").String()
	c.cmdUpdateRepoArg = c.cmdUpdate.Arg("repo name", "Repository to update.").String()
	c.cmdUpdateRevArg = c.cmdUpdate.Arg("revision", "Revision to update.").Int64()

	c.cmdTagRepoArg = c.cmdTag.Arg("repo name", "Repository name.").Required().String()
	c.cmdTagTagArg = c.cmdTag.Arg("tag name", "Tag name.").Required().String()
	c.cmdTagRevArg = c.cmdTag.Arg("revision", "Revision to tag.").Int64()
	c.cmdTagForceFlag = c.cmdTag.Flag("force", "Force tag creation. Default is false.").Short('f').Bool()

	c.cmdDeleteRepoArg = c.cmdDelete.Arg("repo name", "Repository name.").String()

	return c
}

func (c *Cli) Run() error {
	var err error
	c.action = kingpin.MustParse(c.Parse(os.Args[1:]))
	c.app, err = app.NewApp(*c.configFile)
	if err != nil {
		return err
	}

	switch c.action {
	case "create":
		err = c.createCli()
	case "list":
		err = c.listCli()
	case "update":
		err = c.updateCli()
	case "tag":
		err = c.tagCli()
	case "delete":
		err = c.deleteCli()
	}

	return err
}

func (c *Cli) createCli() error {
	c.app.Create(c.action)
	return nil
}

func (c *Cli) listCli() error {
	c.app.List(*c.cmdListRepoArg)
	return nil
}

func (c *Cli) updateCli() error {
	c.app.Update(*c.cmdUpdateRepoArg, *c.cmdUpdateRevArg)
	return nil
}

func (c *Cli) tagCli() error {
	c.app.Tag(*c.cmdTagRepoArg, *c.cmdTagTagArg, *c.cmdTagRevArg, *c.cmdTagForceFlag)
	return nil
}

func (c *Cli) deleteCli() error {
	c.app.Delete(c.action)
	return nil
}
