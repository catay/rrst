[![Go Report Card](https://goreportcard.com/badge/github.com/catay/rrst)](https://goreportcard.com/report/github.com/catay/rrst)

# Remote Repository Sync Tool (rrst)

## What

A tool to sync remote repositories to a local target.  
Currently it only supports repositories using the [RPM XML Metadata format](http://createrepo.baseurl.org/wiki).

This is a first prototype and still lacks some basic features and functionality.

### Features:

 * support for repomd repo's with no authentication requirements
 * support for SUSE SCC repo's which require a **special** form of authentication
 * support download resume for partially downloaded packages
 * environment variable support for storing the secret SCC registration code
 * staged or merged repo support
 * include only selective package patterns for download (no dependency tracking)
 * caches the repomd metadata

## Why 

The main goal of this small project is to learn the Go programming language.

I also required a simple and lightweight tool to download and synchronize SUSE
package repositories from the SUSE Customer Center.  

This tool makes it possible if you have a valid subscription registration code.

## How

The first step is setting up a config file with the repo details.  
A sample YAML config file can be found in the **examples** directory.

Execute the tool with the config file:

```bash
$ rrst -c examples/config.yaml sync
```

Use the help command to list all functionality.

```bash
$ rrst --help
usage: rsst [<flags>] <command> [<args> ...]

Remote Repository Sync Tool

Flags:
      --help                     Show context-sensitive help (also try --help-long and --help-man).
      --version                  Show application version.
  -c, --config="/etc/rsst.yaml"  Path to alternate YAML configuration file.
  -v, --verbose                  Turn on verbose output. Default is verbose turned off.

Commands:
  help [<command>...]
    Show help.

  list
    List repository names and description.

  sync [<repo name>]
    Synchronize remote to local repository sets.

  show
    Show available repository sets.

  clean [<repo name>]
    Cleanup repository cache.
```
