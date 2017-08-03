// Package version provides exported string variables for various
// versioning-related information. It is intended that the values of
// these variables are set at build-time by via the ldflags -X
// argument to the go build command.
package version

var (
	// Version is the version of the Soapbox release, including
	// API and clients.
	Version = "0.1"

	// GitCommit is the short SHA-1 of the current HEAD from which
	// the binary is compiled.
	GitCommit = ""

	// BuildTime is a timestamp of the compilation.
	BuildTime = ""
)
