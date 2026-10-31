//go:build !windows

package paths

import (
	"os"
	"path/filepath"
)

func userDataDir() (string, error) {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return dir, nil
	}
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".local", "share"), nil
	}
	return "", ErrNoDataDir
}
