package cmd

import (
	"github.com/catay/rrst/cmd/app"
	"github.com/catay/rrst/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"path/filepath"
)

type Cli struct {
	*kingpin.Application
	app                  *app.App
	action               string
	configFile           *string
	verbose              *bool
	cmdCreate            *kingpin.CmdClause
	cmdStatus            *kingpin.CmdClause
	cmdList              *kingpin.CmdClause
	cmdUpdate            *kingpin.CmdClause
	cmdTag               *kingpin.CmdClause
	cmdDelete            *kingpin.CmdClause
	cmdDiff              *kingpin.CmdClause
	cmdServer            *kingpin.CmdClause
	cmdTagForceFlag      *bool
	cmdDeleteForceFlag   *bool
	cmdCreateRepoArg     *string
	cmdStatusRepoArg     *string
	cmdListRepoArg       *string
	cmdListTagsOrRevsArg *[]string
	cmdUpdateRepoArg     *string
	cmdUpdateRevArg      *int64
	cmdTagRepoArg        *string
	cmdTagTagArg         *string
	cmdTagRevArg         *int64
	cmdDeleteRepoArg     *string
	cmdDeleteRevArg      *int64
	cmdDiffRepoArg       *string
	cmdDiffTagsOrRevsArg *[]string
	cmdServerPort        *string
}

func NewCli() *Cli {
	c := &Cli{
		Application: kingpin.New(filepath.Base(os.Args[0]), "Remote Repository Sync Tool"),
	}
	c.Version(version.FullVersionString)
	c.Author(version.Author)
	c.configFile = c.Flag("config", "Path to alternate YAML configuration file.").Short('c').Default(app.DefaultConfig).String()
	c.verbose = c.Flag("verbose", "Turn on verbose output. Default is verbose turned off.").Short('v').Bool()
	c.cmdCreate = c.Command("create", "Create custom repositories. **NOT IMPLEMENTED**")
	c.cmdStatus = c.Command("status", "Show status of repositories, revisions and tags.")
	c.cmdList = c.Command("list", "List the packages of a repository.")
	c.cmdUpdate = c.Command("update", "Update repositories with upstream content.")
	c.cmdTag = c.Command("tag", "Tag repository revisions.")
	c.cmdDelete = c.Command("delete", "Delete repository revisions and tags.")
	c.cmdDiff = c.Command("diff", "Show package differences between repository tags.")
	c.cmdServer = c.Command("server", "HTTP server serving repositories.")

	c.cmdCreateRepoArg = c.cmdCreate.Arg("repo name", "Repository name.").String()
	c.cmdStatusRepoArg = c.cmdStatus.Arg("repo name", "Repository name.").String()
	c.cmdListRepoArg = c.cmdList.Arg("repo name", "Repository name.").Required().String()
	c.cmdListTagsOrRevsArg = c.cmdList.Arg("tag|revision", "Show the packages matching a specific set of tags or revisions.").Strings()

	c.cmdUpdateRepoArg = c.cmdUpdate.Arg("repo name", "Repository to update.").String()
	c.cmdUpdateRevArg = c.cmdUpdate.Arg("revision", "Revision to update.").Int64()

	c.cmdTagRepoArg = c.cmdTag.Arg("repo name", "Repository name.").Required().String()
	c.cmdTagTagArg = c.cmdTag.Arg("tag name", "Tag name.").Required().String()
	c.cmdTagRevArg = c.cmdTag.Arg("revision", "Revision to tag.").Required().Int64()
	c.cmdTagForceFlag = c.cmdTag.Flag("force", "Force tag creation. Default is false.").Short('f').Bool()

	c.cmdDeleteRepoArg = c.cmdDelete.Arg("repo name", "Repository name.").Required().String()
	c.cmdDeleteRevArg = c.cmdDelete.Arg("revision", "Revision to delete.").Int64()
	c.cmdDeleteForceFlag = c.cmdDelete.Flag("force", "Force deletion, never prompt. Default is false.").Short('f').Bool()

	c.cmdDiffRepoArg = c.cmdDiff.Arg("repo name", "Repository name.").Required().String()
	c.cmdDiffTagsOrRevsArg = c.cmdDiff.Arg("tag|revision", "Compare package versions between repository tags or revisions.").Required().Strings()

	c.cmdServerPort = c.cmdServer.Flag("port", "Port number to listen on.").Short('p').Default(app.DefaultPort).String()
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
	case "status":
		err = c.statusCli()
	case "list":
		err = c.listCli()
	case "update":
		err = c.updateCli()
	case "tag":
		err = c.tagCli()
	case "diff":
		err = c.diffCli()
	case "delete":
		err = c.deleteCli()
	case "server":
		err = c.serverCli()
	}

	return err
}

func (c *Cli) createCli() error {
	c.app.Create(c.action)
	return nil
}

func (c *Cli) statusCli() error {
	c.app.Status(*c.cmdStatusRepoArg)
	return nil
}

func (c *Cli) listCli() error {
	c.app.List(*c.cmdListRepoArg, *c.cmdListTagsOrRevsArg...)
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

func (c *Cli) diffCli() error {
	c.app.Diff(*c.cmdDiffRepoArg, *c.cmdDiffTagsOrRevsArg...)
	return nil
}

func (c *Cli) deleteCli() error {
	c.app.Delete(*c.cmdDeleteRepoArg, *c.cmdDeleteRevArg, *c.cmdDeleteForceFlag)
	return nil
}

func (c *Cli) serverCli() error {
	return c.app.Server(*c.cmdServerPort)
}
