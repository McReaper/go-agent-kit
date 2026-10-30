package paths_test

import (
	"path/filepath"
	"testing"

	"example.com/notes/internal/platform/paths"
)

func TestDataDirIsUnderTheUserDataHome(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", filepath.FromSlash("/data"))
	got, err := paths.DataDir("notes")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(filepath.FromSlash("/data"), "notes"); got != want {
		t.Errorf("DataDir = %q, want %q", got, want)
	}
}
