* Release x.x.x (xxxx/xx/xx)
  - Implement the delete command.
  - (#22) Change default config location to /etc/rrst/config.yaml.
  - Add support for both the createrepo (Python) and createrepo_c (C) command.
  - Migrate dependency management to Go Modules.
* Release 0.4.0 (2019/08/04)
  - The functionality of the list command is replaced by the status command.
  - ( #9) Implement a list and diff command to show package version info.  
  - (#17) Make the revision parameter mandatory with the tag command.
* Release 0.3.0 (2019/05/08)
  - Improve versioning and build system.
  - Take into account deleted packages on the local filesystem.
  - Implement functionality for local only repositories.
  - Store the temporary SCC json file in os.TempDir().
  - Fix the date formatting in the list commmand.
* Release 0.2.0 (2019/01/09)
  - Complete redesign of the tool.
  - Built-in web server to serve RPM's.
  - Implement staging through revisions and tags.
* Release 0.1.0 (2018/08/05)
  - Support for repomd repo's with no authentication requirements.
  - Support for SUSE SCC repo's which require a special form of authentication.
  - Support download resume for partially downloaded packages.
  - Environment variable support for storing the secret SCC registration code.
  - HTTP(s) proxy support
  - Staged or merged repo support.
  - Include only selective package patterns for download (no dependency tracking).
  - Caches the repomd metadata.
