* Release 0.2.0 (2019/01/09)
  - complete redesign of the tool
  - built-in web server to serve RPM's
  - implement staging through revisions and tags.
* Release 0.1.0 (2018/08/05)
  - support for repomd repo's with no authentication requirements
  - support for SUSE SCC repo's which require a special form of authentication
  - support download resume for partially downloaded packages
  - environment variable support for storing the secret SCC registration code
  - HTTP(s) proxy support
  - staged or merged repo support
  - include only selective package patterns for download (no dependency tracking)
  - caches the repomd metadata
