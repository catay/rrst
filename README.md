[![CircleCI](https://circleci.com/gh/catay/rrst.svg?style=svg)](https://circleci.com/gh/catay/rrst)
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
 * HTTP(s) proxy support
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

## Build

The binary can be easily build inside a Docker container if Go is not installed on the local system.  
Note that the binary will only work on recent x86_64 Linux architectures.

Clone the source code from GitHub.

```bash
$ git clone git@github.com:catay/rrst.git && cd rrst
```
Build the image from the Dockerfile.

```bash
$ docker build -t rrst .
```

Spin up the container based on the image previously created.

```bash
docker run rrst rrst --version
```

Copy the binary from the container to the local filesystem path `/var/tmp/rrst`.  
There are better ways on how to do this, but this will avoid any SELinux hassle with shared volumes.

Replace in the below snippets `<CONTAINER ID>` with the id provided in `docker container ls -a`.

```bash
$ docker cp <CONTAINER ID>:/go/src/app/rrst /var/tmp/rrst
```

Remove the container.

```bash
$ docker rm <CONTAINER ID>
```

Test if the binary works on your local system.

```bash
$ /var/tmp/rrst -c examples/config.yaml list
Configured repositories:
   SLES-12-3-X86_64-updates
   SLES-11-4-X86_64-updates
   RHEL-7-4-X86_64-os
   SLES-12-3-X86_64-opensuse
```

