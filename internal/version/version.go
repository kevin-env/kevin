// Package version holds kevin's version string, embedded from a file the
// release process writes and commits before tagging, and the commit the
// binary was built from. internal/cmd and
// internal/relay both read it - internal/relay can't import internal/cmd
// without an import cycle (internal/cmd imports internal/engine, which
// imports internal/relay) - so it lives here instead, as the one source
// both derive from.
package version

import (
	_ "embed"
	"runtime/debug"
	"strings"
)

//go:embed VERSION
var versionFile string

// String is kevin's version.
var String = strings.TrimSpace(versionFile)

// shortCommitLen is how many characters of the commit hash buildCommit keeps.
const shortCommitLen = 7

// buildCommit returns the short commit hash the binary was built from, with
// "+dirty" appended when the working tree had uncommitted changes. It is
// empty when the build carries no VCS information.
func buildCommit() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	return commitFrom(info)
}

// Full returns [String] followed by the build's commit in parentheses, or
// [String] alone when the build carries no commit.
func Full() string {
	return format(String, buildCommit())
}

func commitFrom(info *debug.BuildInfo) string {
	var rev string
	var dirty bool
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			rev = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}
	if rev == "" {
		return ""
	}
	rev = rev[:min(len(rev), shortCommitLen)]
	if dirty {
		rev += "+dirty"
	}
	return rev
}

func format(version, commit string) string {
	if commit == "" {
		return version
	}
	return version + " (" + commit + ")"
}
