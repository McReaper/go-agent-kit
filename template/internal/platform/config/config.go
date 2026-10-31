// Package config reads the app's settings and checks them at startup: a
// bad setting stops the app before it does anything, with a message that
// names the setting and what it takes.
package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/McReaper/go-agent-kit/template/internal/platform/paths"
)

// Config is the app's checked settings.
type Config struct {
	// DataDir is where the app keeps its data and its journal.
	DataDir string
}

// Load reads app's settings from getenv: <APP>_DATA_DIR, an absolute path,
// by default the app's directory under the system's user data directory.
func Load(app string, getenv func(string) string) (Config, error) {
	key := strings.ToUpper(app) + "_DATA_DIR"
	dir := getenv(key)
	if dir == "" {
		d, err := paths.DataDir(app)
		if err != nil {
			return Config{}, fmt.Errorf("%s unset: %w", key, err)
		}
		return Config{DataDir: d}, nil
	}
	if !filepath.IsAbs(dir) {
		return Config{}, fmt.Errorf("%s=%q: give an absolute path", key, dir)
	}
	return Config{DataDir: filepath.Clean(dir)}, nil
}
