package config_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/McReaper/go-agent-kit/template/internal/platform/config"
)

func env(vars map[string]string) func(string) string {
	return func(k string) string { return vars[k] }
}

func TestTheDataDirComesFromTheAppsVariable(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "x", "..", "data")
	c, err := config.Load("notes", env(map[string]string{"NOTES_DATA_DIR": dir}))
	if err != nil {
		t.Fatal(err)
	}
	if c.DataDir != filepath.Clean(dir) {
		t.Errorf("DataDir = %q, want %q", c.DataDir, filepath.Clean(dir))
	}
}

func TestARelativeDataDirStopsTheApp(t *testing.T) {
	_, err := config.Load("notes", env(map[string]string{"NOTES_DATA_DIR": "data"}))
	if err == nil || !strings.Contains(err.Error(), "NOTES_DATA_DIR") || !strings.Contains(err.Error(), "absolute") {
		t.Errorf("err = %v, want one naming NOTES_DATA_DIR and asking for an absolute path", err)
	}
}

func TestTheDataDirDefaultsToTheSystemsOwn(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_DATA_HOME", base)
	t.Setenv("LOCALAPPDATA", base)
	c, err := config.Load("notes", env(nil))
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(base, "notes"); c.DataDir != want {
		t.Errorf("DataDir = %q, want %q", c.DataDir, want)
	}
}
