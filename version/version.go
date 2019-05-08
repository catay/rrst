package version

// The version string will be injected during the build.
var Version = "x.x.x"

// The commit string will be injected during the build.
var Commit = "unknown"

var FullVersionString = "rrst " + Version + " (" + Commit + ")"

// Global constants.
const (
	Author = "Steven Mertens <steven.mertens@catay.be>"
)
