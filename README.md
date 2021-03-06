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
  * [rrst status](#rrst-status)
  * [rrst list](#rrst-list)
  * [rrst update](#rrst-update)
  * [rrst tag](#rrst-tag)
  * [rrst delete](#rrst-delete)
  * [rrst diff](#rrst-diff)
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
$ wget https://github.com/catay/rrst/releases/download/0.4.0/rrst-0.4.0.tar.gz
```

Extract the archive which contain the rrst binary.

```bash
$ tar xzf rrst-0.4.0.tar.gz
```

Test out the unpacked rrst binary.

```bash
$ ./rrst --version
rrst 0.4.0 (1ae069f92073594f557cbb7af09f5689ff9df797)
```

RPM or DEB packages will be provided in the future.

### Build from source

There are a few prerequisites when building from source:

* Install the [Go tools](https://golang.org/doc/install).
* Install [GNU Make](https://www.gnu.org/software/make/).

1. Clone the upstream rrst Git repository.

   ```bash
   $ git clone https://github.com/catay/rrst.git
   ```

2. Run `make` to build the rrst tool.

   ```bash
   $ cd rrst && make
   ```
3. Test the build.

   ```bash
   $ ./rrst --version
   rrst x.x.x (8159eb12e14f0e4820d751e3535c34947861aa4c)
   ```

> **Note:** An unofficial release build will have a x.x.x version.

### Build with Docker

The binary can also be build inside a Docker container if Go is not 
installed on the local system.

1. Clone the source code from GitHub.

   ```bash
   $ git clone git@github.com:catay/rrst.git && cd rrst
   ```
2. Build the image from the Dockerfile.

   ```bash
   $ docker build -t rrst .
   ```

3. Spin up the container based on the image previously created.

   ```bash
   docker run rrst rrst --version
   ```
4. Copy the binary from the container to the local filesystem.  

   ```bash
   $ docker cp <CONTAINER ID>:/go/src/app/rrst /var/tmp/rrst

5. Remove the container.

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
      --help     Show context-sensitive help (also try --help-long and --help-man).
      --version  Show application version.
  -c, --config="/etc/rrst/config.yaml"
                 Path to alternate YAML configuration file.
  -v, --verbose  Turn on verbose output. Default is verbose turned off.

Commands:
  help [<command>...]
    Show help.

  create [<repo name>]
    Create custom repositories. **NOT IMPLEMENTED**

  status [<repo name>]
    Show status of repositories, revisions and tags.

  list <repo name> [<tag|revision>...]
    List the packages of a repository.

  update [<repo name>] [<revision>]
    Update repositories with upstream content.

  tag [<flags>] <repo name> <tag name> <revision>
    Tag repository revisions.

  delete [<flags>] <repo name> [<revision>]
    Delete repository revisions and tags.

  diff <repo name> <tag|revision>...
    Show package differences between repository tags.

  server [<flags>]
    HTTP server serving repositories.
```

### rrst create

**Not implemented**. 
This subcommand will make it possible to create a local repository not tied to a remote one.

### rrst status

The status command without arguments shows the status of the configured repositories.

```bash
$ rrst -c config.yaml status
ID    REPOSITORY                   ENABLED    #REVISIONS    #TAGS    UPDATED
1     CENTOS-7-6-X86_64-updates    true       5             5        2019-08-02 14:25:00
2     SLES-15-0-X86_64-updates     true       2             2        2019-08-04 11:52:39
3     DUMMY-1-0-X86_64-updates     true       1             1        2019-01-21 20:16:22
```

Providing the repository name as argument will show more detailed information
about the available revisions and tags.

```bash
$ rrst -c config.yaml status CENTOS-7-6-X86_64-updates
REVISION      CREATED                TAGS
1546898144    2019-01-07 22:55:44    prd 
1547291211    2019-01-12 12:06:51    tst
1547670979    2019-01-16 21:36:19    dev
1547680849    2019-01-17 00:20:49    latest
```

### rrst list

The list command shows the packages of a repository who are part of a set of tags or revisions.

Example, list the packages of the **prd** and **latest** tags.

```bash
$ rrst -c config.yaml list CENTOS-7-6-X86_64-updates prd latest
PACKAGE                                            prd                         latest
NetworkManager-wifi.x86_64                         1.12.0-8.el7_6              1.12.0-8.el7_6
fence-agents-drac5.x86_64                          4.2.1-11.el7_6.1            4.2.1-11.el7_6.1
pcp-pmda-docker.x86_64                             4.1.0-5.el7_6               4.1.0-5.el7_6
libvncserver.x86_64                                -                           0.9.9-13.el7_6
pcp-export-pcp2zabbix.x86_64                       4.1.0-5.el7_6               4.1.0-5.el7_6
libvncserver-devel.i686                            -                           0.9.9-13.el7_6
cronie-anacron.x86_64                              1.4.11-20.el7_6             1.4.11-20.el7_6
libguestfs-xfs.x86_64                              1.38.2-12.el7_6.1           1.38.2-12.el7_6.1
xorg-x11-server-Xvfb.x86_64                        1.20.1-5.el7                1.20.1-5.el7
...
```

It is also possible to list the packages of both tags and revisions.

```bash
$ rrst -c config.yaml list CENTOS-7-6-X86_64-updates 1547291211 dev
PACKAGE                                            1547291211                  dev
NetworkManager-glib-devel.i686                     1.12.0-8.el7_6              1.12.0-8.el7_6
systemd-libs.x86_64                                -                           219-62.el7_6.2
ruby-libs.x86_64                                   2.0.0.648-34.el7_6          2.0.0.648-34.el7_6
pcp.x86_64                                         4.1.0-5.el7_6               4.1.0-5.el7_6
java-1.7.0-openjdk-demo.x86_64                     1.7.0.201-2.6.16.1.el7_6    1.7.0.201-2.6.16.1.el7_6
java-1.8.0-openjdk-devel.x86_64                    1.8.0.191.b12-1.el7_6       1.8.0.191.b12-1.el7_6
ghostscript-devel.i686                             9.07-31.el7_6.6             9.07-31.el7_6.6
pacemaker-remote.x86_64                            1.1.19-8.el7_6.2            1.1.19-8.el7_6.2
pcp-webapi.x86_64                                  4.1.0-5.el7_6               4.1.0-5.el7_6
...
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

The delete command deletes revisions and associated tags.
For now the command only deletes the metadata and not the content.

Delete all revisions of a repository.

```bash
$ rrst -c config.yaml delete CENTOS-7-6-X86_64-updates
```

Delete a specific revision of a repository.

```bash
$ rrst -c config.yaml delete CENTOS-7-6-X86_64-updates 1546898144
```

### rrst diff

The diff command compares package versions between tags or revisions.
It takes a repository name and a whitespace delimited list of tags or revisions.


```bash
$ rrst -c config.yaml diff CENTOS-7-6-X86_64-updates production test latest
PACKAGE                           production     test                      latest
libvncserver.i686                 -              -                         0.9.9-13.el7_6
systemd.x86_64                    -              219-62.el7_6.2            219-62.el7_6.2
elinks.x86_64                     -              0.12-0.37.pre6.el7.0.1    0.12-0.37.pre6.el7.0.1
systemd-libs.x86_64               -              219-62.el7_6.2            219-62.el7_6.2
libvncserver.x86_64               -              -                         0.9.9-13.el7_6
libgudev1.i686                    -              219-62.el7_6.2            219-62.el7_6.2
tzdata.noarch                     2018g-1.el7    2018i-1.el7               2018i-1.el7
tzdata-java.noarch                2018g-1.el7    2018i-1.el7               2018i-1.el7
systemd-python.x86_64             -              219-62.el7_6.2            219-62.el7_6.2
systemd-devel.i686                -              219-62.el7_6.2            219-62.el7_6.2
systemd-networkd.x86_64           -              219-62.el7_6.2            219-62.el7_6.2
libgudev1.x86_64                  -              219-62.el7_6.2            219-62.el7_6.2
libgudev1-devel.x86_64            -              219-62.el7_6.2            219-62.el7_6.2
systemd-sysv.x86_64               -              219-62.el7_6.2            219-62.el7_6.2
systemd-devel.x86_64              -              219-62.el7_6.2            219-62.el7_6.2
libvncserver-devel.x86_64         -              -                         0.9.9-13.el7_6
systemd-resolved.i686             -              219-62.el7_6.2            219-62.el7_6.2
libgudev1-devel.i686              -              219-62.el7_6.2            219-62.el7_6.2
systemd-journal-gateway.x86_64    -              219-62.el7_6.2            219-62.el7_6.2
systemd-libs.i686                 -              219-62.el7_6.2            219-62.el7_6.2
systemd-resolved.x86_64           -              219-62.el7_6.2            219-62.el7_6.2
libvncserver-devel.i686           -              -                         0.9.9-13.el7_6
```

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
* Add a credits file
* Move all the repository management server-side and provide a REST API
* Add and complete Go (Ginkgo) tests
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

