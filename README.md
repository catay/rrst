[![CircleCI](https://circleci.com/gh/catay/rrst.svg?style=svg)](https://circleci.com/gh/catay/rrst)
[![Go Report Card](https://goreportcard.com/badge/github.com/catay/rrst)](https://goreportcard.com/report/github.com/catay/rrst)

# Remote Repository Sync Tool (rrst)

Use rrst to sync remote rpm-md style 
([RPM XML Metadata format](http://createrepo.baseurl.org/wiki)) 
repositories to a local target.

The tool will detect remote repository metadata changes and link this to revisions.
Revisions can be tagged and those tags can be served as a RPM repository with the 
build-in webserver to local clients.

This allows you to implement a [DTAP](https://en.wikipedia.org/wiki/Development,_testing,_acceptance_and_production)
strategy for RPM repositories.

The key goal of this tool is to keep it very lightweight and provide everything through one binary.

## Table of Contents

* [Installation](#installation)
  * [Precompiled binaries](#precompiled-binaries)
  * [Build from source](#build-from-source)
  * [Build with Docker](#build-with-docker)
* [How it works](#how-it-works)
* [Configuration reference](#configuration-reference)
  * [global](#global)
  * [providers](#providers)
    * [SUSE](#suse)
  * [repositories](#repositories)
* [Command reference](#command-reference)
  * [rrst help](#rrst-help)
  * [rrst create](#rrst-create)
  * [rrst list](#rrst-list)
  * [rrst update](#rrst-update)
  * [rrst tag](#rrst-tag)
  * [rrst delete](#rrst-delete)
  * [rrst server](#rrst-server)
* [Design](#design)
* [Roadmap](#roadmap)
* [License](#license)

## Installation

The rrst tool is only supported and tested on Linux 64-bit based distro's.
There are a two ways to install rrst on a system.

### Precompiled binaries

Precompiled binaries are available for each release at the
[release](https://github.com/catay/rrst/releases) section.

Fetch the archive.

```bash
$ wget https://github.com/catay/rrst/releases/download/0.2.0/rrst-0.2.0.tar.gz
```

Extract the archive which contain the rrst binary.

```bash
$ tar xzf rrst-0.2.0.tar.gz
```

RPM or DEB packages will be provided in the future.

### Build from source

Compiling the binary requires setting up the [Go tools](https://golang.org/doc/install).

We will cover a quick standard Go install into `${HOME}` to get rrst build from source.

Download the latest Go binary release in `${HOME}`.

```bash
$ wget https://dl.google.com/go/go1.11.4.linux-amd64.tar.gz
```

Extract the tarball into the `tmp` directory. 

```bash
$ tar -C /var/tmp -xzf go1.11.4.linux-amd64.tar.gz
```

Create a Go workspace directory in `${HOME}` and copy the binary.

```bash
$ mkdir -p ${HOME}/go/{bin,src}
$ cp /var/tmp/go/bin/go ${HOME}/go/bin
```

Add the Go binary path to `${PATH}`.

```bash
$ export PATH=${PATH}:${HOME}/go/bin
```

We also require [dep](https://golang.github.io/dep/), the Go dependency management tool.

Download the latest dep binary release in `${HOME}`.
This will also automatically move the binary into `${HOME}/go/bin`.

```bash
$ curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
```

We are all set now to build the rrst tool.
The first step is to download the source code in the right spot. 

```bash
$ cd ${HOME}/go/src
$ git clone https://github.com/catay/rrst.git
```

Run `make` to build the rrst tool.
This might take a while as it will also fetch all dependencies.

```bash
$ cd ${HOME}/go/src/rrst && make
```

A successful build results in a rrst binary in the current directory.

```bash
$ ./rrst --version
0.1.0
```

### Build with Docker

The binary can also be build inside a Docker container if Go is not 
installed on the local system.

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

## How it works

Let's take a real use case as example to see rrst in action.

Imagine we have an on premise environment consisting out of many CentOS 
servers.

On a regular basis we want to update the packages on all those servers as 
required by the internal patch management policy. 

We can point each of those servers to a remote CentOS mirror site but 
that's not really a good practise for various reasons.

 * might result in different package versions installed across servers
 * makes you dependent on remote infrastructure out of your control
 * no possiblities to have a testing vs production repository 
 * waste of internet bandwidth by downloading the same packages multiple times

The solution to the above is hosting a package repository on premise. 

The first step would be to mirror the CentOS update repository on to 
an internal server.

Let's assume we already installed the rrst binary on that internal
server and now want to configure it to mirror the upstream CentOS 
update repository.

Fire up your favorite editor and create a file called `config.yaml`
with the below content.

```bash
global:
  content_path: /var/cache/rrst
repositories:
  - id: 1
    name: CENTOS-7-6-X86_64-updates
    type: rpmmd
    enabled: true
    remote_uri: http://ftp.belnet.be/mirror/ftp.centos.org/7.6.1810/updates/x86_64/
    content_suffix_path: CENTOS/7/6/1810/x86_64/updates
```

The above is all it takes to configure fetching packages from a remote
CentOS update site.

We now have a very basic configuration file which we can use in combination with rrst.

The list command provides information of the configured repositories.

```bash
$ rrst -c config.yaml list
ID    REPOSITORY                   ENABLED    #REVISIONS    #TAGS    UPDATED
1     CENTOS-7-6-X86_64-updates    true       0             0        never
```

The above output shows the CENTOS-7-6-X86_64-updates repository is
enabled and has never been updated from remote.

We can fetch all the remote packages by issuing the update command.
A counter will display a basic package download status.

```bash
$ rrst -c config.yaml update
CENTOS-7-6-X86_64-updates                       [  112/625  ]   Packages/firefox-60.3.0-1.el7.centos.i686.rpm
```

When all packages are downloaded it will display this like below. 
The update command is restartable in case of interuption and will pick up where it left off.

```bash
$ rrst -c config.yaml update
CENTOS-7-6-X86_64-updates                       [  625/625  ]   Done
```

The list command will now give a different output.

```bash
$ rrst -c config.yaml list
ID    REPOSITORY                   ENABLED    #REVISIONS    #TAGS    UPDATED
1     CENTOS-7-6-X86_64-updates    true       1             1        2019-1-7 22:55:44
```

The output now shows we have one revision, one tag and also displays the
timestamp of the last update.

We can get more detailed information with the list command when we specifiy the repo
name as argument.

```bash
$ rrst -c config.yaml list CENTOS-7-6-X86_64-updates
REVISIONS     CREATED              TAGS
1546898144    2019-1-7 22:55:44    latest
```

A *revision* is a snapshot of the state of the repository when it was downloaded.
In case the remote repository changes and we issue a new update,  a new
revision will be created capturing the current state.

The *latest* tag is automatically created and will always link to the latest revision.

It is also possible to create custom tags linked to a specific revision.

For example we can assign the production tag to revision 1546898144.

```bash
$ rrst -c config.yaml tag CENTOS-7-6-X86_64-updates production 1546898144
```

This makes it possible to use tags to link a certain revision to a 
specific dev, test or production environment.

```bash
$ rrst -c config.yaml list CENTOS-7-6-X86_64-updates
REVISIONS     CREATED              TAGS
1546898144    2019-1-7 22:55:44    latest, production
```

The real power of tags comes into play when the build-in webserver is
used to serve repositories across clients.

```bash
$ rrst -c config.yaml server
2019/01/08 00:44:59 Start server: :4280
2019/01/08 00:44:59 register latest url: /CENTOS/7/6/1810/x86_64/updates/latest/
2019/01/08 00:44:59 register production url: /CENTOS/7/6/1810/x86_64/updates/production/
```

The webserver will make each tag available on a different URL, which
under the hood points to the linked revision, which points on its turn to a 
specific state of the repository.

Tags allow you to implement a [DTAP](https://en.wikipedia.org/wiki/Development,_testing,_acceptance_and_production)
approach for RPM repositories.

## Configuration reference

This section will cover in more detail the different configuration options.

### global

The global section contains the key/values that apply program wide.

|key             |value |description |
|----------------|------|------------|
|content_path    |string|The parent path where rrst will store all the downloaded packages and metadata.|
|max_revs_to_keep|string|Maximum revisions to keep with no tags linked. (**not implemented**)| 
|providers       |array|Provider specific configuration for vendor repositories like authentication.| 

### providers

The providers section is a first try implementation to configure provider/vendor specific settings.
Currently it only supports the [SUSE Connect API](https://scc.suse.com/connect/v4/documentation).

#### SUSE

To download packages for SUSE Linux Enterprise Server you require a valid subscription.
This will give you access to the [SUSE Customer Center](https://scc.suse.com/login).

The rrst tool supports downloading packages from SCC. You only require a valid registration code.
You can find the registration code in the subscription information section. 
The *Base and Extension Products* section shows what the subscription covers and the repository URL's.

Config file example for a SUSE SCC repository.

```bash
global:
  content_path: /var/tmp/rrst
  providers:
    - id: SLES01
      provider: SUSE
      variables:
        - name: scc_reg_code
          value: ${SCC_REG_CODE_01}
repositories:
  - id: 1
    name: SLES-15-0-X86_64-updates
    type: rpmmd
    provider_id: SLES01
    enabled: true
    remote_uri: https://updates.suse.com/SUSE/Updates/SLE-Product-SLES/15/x86_64/update
    content_suffix_path: SLES/15/0/x86_64/updates
```

The providers key takes an array of providers. A provider is referenced through a unique id. 
This id can be freely chosen and it allows multiple configs with different variables for the 
same provider.

In the SUSE case we only need to specify the variable `scc_reg_code` with a valid registration
code as value. In the above case we point to an environment variable `${SCC_REG_CODE_01}` that
holds the code.

### repositories

The repositories section contains a list of all the repository configurations.

|key         |value          |description|
|------------|---------------|------------|
|id|integer|A integer id value for the repository. Will probably be removed.|
|name|string|The short name of the repository.|
|type|string|The type of the repository. For now only rpm-md is supported.|
|provider_id|string|The provider id to map with.|
|enabled|boolean|Enable or disable the repository. Values are true or false.|
|remote_uri|string|The URL of the remote repository containing the repodata directory.|
|content_suffix_path|string|Extension of the content_path where the packages will be stored and served from.|


## Command reference

An overview of the `rrst` subcommands.

### rrst help

Shows the general help page of `rrst` tool. 
Help for a subcommand can be obtained through `rrst help <command>`.

```bash
$ rrst help
usage: rrst [<flags>] <command> [<args> ...]

Remote Repository Sync Tool

Flags:
      --help                     Show context-sensitive help (also try --help-long and --help-man).
      --version                  Show application version.
  -c, --config="/etc/rrst.yaml"  Path to alternate YAML configuration file.
  -v, --verbose                  Turn on verbose output. Default is verbose turned off.

Commands:
  help [<command>...]
    Show help.

  create [<repo name>]
    Create custom repositories. **NOT IMPLEMENTED**

  list [<repo name>]
    List repositories, revisions and tags.

  update [<repo name>] [<revision>]
    Update repositories with upstream content.

  tag [<flags>] <repo name> <tag name> [<revision>]
    Tag repository revisions.

  delete [<repo name>]
    Delete repositories. **NOT IMPLEMENTED**

  server [<flags>]
    HTTP server serving repositories.
```

### rrst create

**Not implemented**. 
This subcommand will make it possible to create a local repository not tied to a remote one.

### rrst list

The list command without arguments shows the status of the configured repositories.

```bash
$ rrst -c config.yaml list
ID    REPOSITORY                   ENABLED    #REVISIONS    #TAGS    UPDATED
1     CENTOS-7-6-X86_64-updates    true       1             2        2019-1-7 22:55:44
```

Providing the repository name as argument will show more detailed information
about the available revisions and tags.

```bash
$ rrst -c config.yaml list CENTOS-7-6-X86_64-updates
REVISIONS     CREATED              TAGS
1546898144    2019-1-7 22:55:44    latest, production
```

### rrst update

The update command with no arguments will mirror the packages from all
the configured remote repositories to the local target.

```bash
$ rrst -c config.yaml update
CENTOS-7-6-X86_64-updates                       [  112/625  ]   Packages/firefox-60.3.0-1.el7.centos.i686.rpm
```

You can also update a specific repository by providing the name as argument.

```bash
$ rrst -c config.yaml update CENTOS-7-6-X86_64-updates
CENTOS-7-6-X86_64-updates                       [  625/625  ]   Done
```

Check also `rrst help update` for more options.

### rrst tag

The tag subcommand creates tags linked to repository revisions. 
It takes a repository name, tag name and revision as arguments. 

A tag can only contain letters, digits and underscores. 

```bash
$ rrst -c config.yaml tag CENTOS-7-6-X86_64-updates my_custom_tag 1546898144
```

### rrst delete

**Not implemented**. 
This subcommand will delete repositories, revisions and tags.

### rrst server

The server command starts a basic webserver on port 4280.
It will serve packages for each tag of a repository.

```bash
$ rrst -c config.yaml server
2019/01/08 22:22:45 Start server: :4280
2019/01/08 22:22:45 register latest url: /CENTOS/7/6/1810/x86_64/updates/latest/
2019/01/08 22:22:45 register production url: /CENTOS/7/6/1810/x86_64/updates/production/
```

The port number can be changed with the -p flag. See `rrst help server` for more details.

## Design

To be completed.

## Roadmap

There are a lot of features and improvements possible in the current
implementation.

* Provide rrst RPM packages for the main Linux distributions
* Switch to [version 4 UUID's](https://en.wikipedia.org/wiki/Universally_unique_identifier#Version_4_(random)) to track revisions in the filesystem
* Implement a locking mechanism when a repository update is in progress
* Add metalink or mirrorlist support
* Set HTTP user agent to a custom string
* Implement diff functionality between two revisions
* Add a credits file
* Move all the repository management server-side and provide a REST API
* Add and complete Go (Ginkgo) tests
* Inject commit hash into the version string
* Add delta RPM support
* Replace content_suffix_path by the repo name (?)
* Remove repository id as an array is ordered anyway
* Verify the checksum hash of the packages
* HTTPS support for the webserver
* Parallel package and repository downloads
* Check if there is enough free capacity on the filesystem before download
* GPG support
* Add a quiet cli flag

## License

Apache License 2.0, see [LICENSE](https://github.com/catay/rrst/blob/master/LICENSE.md).

