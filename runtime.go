package logging

import "runtime/debug"

const (
	// BuildCommitChars is the number of characters to show in the build commit.
	BuildCommitChars = 7
)

var (
	// buildCommit is the git source commit (only first BuildCommitChars characters).
	// Use linker flag to customize it: -X 'github.com/osbuild/logging.buildCommit=1234567
	buildCommit string

	// buildTime is the build time, if VCS is available
	// Use linker flag to customize it: -X 'github.com/osbuild/logging.buildTime=2021-01-01T00:00:00Z'
	buildTime string
)

func init() {
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, bs := range bi.Settings {
			switch bs.Key {
			case "vcs.revision":
				if len(bs.Value) > BuildCommitChars {
					buildCommit = bs.Value[0:BuildCommitChars]
				}
			case "vcs.time":
				buildTime = bs.Value
			}
		}
	}
}

// BuildID returns the build ID, typically a git commit but can be overriden via a linker flag.
// This is the short version, up to BuildCommitChars characters long.
func BuildID() string {
	return buildCommit
}

// BuildTime returns the build time, if available.
func BuildTime() string {
	return buildTime
}
