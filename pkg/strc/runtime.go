package strc

import "runtime/debug"

const (
	// BuildCommitChars is the number of characters to show in the build commit.
	BuildCommitChars = 7
)

var (
	// BuildCommit is the git source commit (only first BuildCommitChars characters)
	BuildCommit string

	// BuildTime represents build date and time
	BuildTime string

	// BuildGoVersion carries Go version the binary was built with
	BuildGoVersion string
)

func init() {
	BuildTime = "N/A"
	BuildCommit = "HEAD"

	if bi, ok := debug.ReadBuildInfo(); ok {
		BuildGoVersion = bi.GoVersion

		for _, bs := range bi.Settings {
			switch bs.Key {
			case "vcs.revision":
				if len(bs.Value) > BuildCommitChars {
					BuildCommit = bs.Value[0:BuildCommitChars]
				}
			case "vcs.time":
				BuildTime = bs.Value
			}
		}
	}
}
