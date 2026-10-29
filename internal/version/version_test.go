package version

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommitFrom(t *testing.T) {
	settings := func(kv ...string) *debug.BuildInfo {
		info := &debug.BuildInfo{}
		for i := 0; i < len(kv); i += 2 {
			info.Settings = append(info.Settings, debug.BuildSetting{Key: kv[i], Value: kv[i+1]})
		}
		return info
	}

	t.Run("shortens the revision", func(t *testing.T) {
		got := commitFrom(settings("vcs.revision", "2171a43a0123456789abcdef", "vcs.modified", "false"))
		assert.Equal(t, "2171a43", got)
	})

	t.Run("marks a dirty tree", func(t *testing.T) {
		got := commitFrom(settings("vcs.revision", "2171a43a0123456789abcdef", "vcs.modified", "true"))
		assert.Equal(t, "2171a43+dirty", got)
	})

	t.Run("keeps a revision shorter than the cut", func(t *testing.T) {
		assert.Equal(t, "abc", commitFrom(settings("vcs.revision", "abc")))
	})

	t.Run("is empty without VCS information", func(t *testing.T) {
		assert.Empty(t, commitFrom(settings("-ldflags", "-s -w")))
	})
}

func TestFormat(t *testing.T) {
	assert.Equal(t, "v0.1.0 (2171a43)", format("v0.1.0", "2171a43"))
	assert.Equal(t, "v0.1.0", format("v0.1.0", ""))
}
