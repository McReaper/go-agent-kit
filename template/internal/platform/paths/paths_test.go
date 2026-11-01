package paths_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/McReaper/go-agent-kit/template/internal/platform/paths"
)

func TestTheDataDirIsTheAppsDirectoryUnderTheSystemsOwn(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_DATA_HOME", base)
	t.Setenv("LOCALAPPDATA", base)
	got, err := paths.DataDir("notes")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(base, "notes"); got != want {
		t.Errorf("DataDir = %q, want %q", got, want)
	}
}

func TestNoDataDirWhenTheSystemNamesNone(t *testing.T) {
	for _, v := range []string{"XDG_DATA_HOME", "HOME", "LOCALAPPDATA"} {
		t.Setenv(v, "")
	}
	if _, err := paths.DataDir("notes"); !errors.Is(err, paths.ErrNoDataDir) {
		t.Errorf("err = %v, want ErrNoDataDir", err)
	}
}
