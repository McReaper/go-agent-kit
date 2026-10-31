// Package paths says where the app keeps its data on each system.
package paths

import (
	"errors"
	"path/filepath"
)

// ErrNoDataDir is returned when the system names no place for user data.
var ErrNoDataDir = errors.New("no user data directory: set the environment variable the system uses for it")

// DataDir returns the app's data directory, named app, under the system's
// directory for user data. It doesn't create it.
func DataDir(app string) (string, error) {
	base, err := userDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, app), nil
}
