package logging

import "runtime/debug"

const (
	// BuildCommitChars is the number of characters to show in the build commit.
	BuildCommitChars = 7
)

var (
	// BuildCommit is the git source commit (only first BuildCommitChars characters)
	BuildCommit string

	// BuildCustom overrides the build commit with a custom string
	BuildCustom string
)

func init() {
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, bs := range bi.Settings {
			switch bs.Key {
			case "vcs.revision":
				if len(bs.Value) > BuildCommitChars {
					BuildCommit = bs.Value[0:BuildCommitChars]
				}
			}
		}
	}
}

func BuildID() string {
	if BuildCustom != "" {
		return BuildCustom
	}
	return BuildCommit
}
