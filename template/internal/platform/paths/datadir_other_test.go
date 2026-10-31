//go:build !windows

package paths_test

import (
	"path/filepath"
	"testing"

	"github.com/McReaper/go-agent-kit/template/internal/platform/paths"
)

func TestWithoutXDGDataHomeTheDataDirIsUnderHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("HOME", home)
	got, err := paths.DataDir("notes")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, ".local", "share", "notes"); got != want {
		t.Errorf("DataDir = %q, want %q", got, want)
	}
}
