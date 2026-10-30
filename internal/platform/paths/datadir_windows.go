package paths

import "os"

// WHY: XDG_DATA_HOME is honoured on Windows too, so the override works the
// same on every system; LOCALAPPDATA is the Windows default.
func userDataDir() (string, error) {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return dir, nil
	}
	if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
		return dir, nil
	}
	return "", ErrNoDataDir
}
