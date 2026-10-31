package paths_test

import (
	"path/filepath"
	"testing"

	"example.com/notes/internal/platform/paths"
)

func TestDataDirIsUnderTheUserDataHome(t *testing.T) {
	// WHY: each system reads its own variable (Windows: LOCALAPPDATA; others:
	// XDG_DATA_HOME), so the test sets both to the same base.
	t.Setenv("XDG_DATA_HOME", filepath.FromSlash("/data"))
	t.Setenv("LOCALAPPDATA", filepath.FromSlash("/data"))
	got, err := paths.DataDir("notes")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(filepath.FromSlash("/data"), "notes"); got != want {
		t.Errorf("DataDir = %q, want %q", got, want)
	}
}
