// Copyright 2026 Notes Inc. All rights reserved.

package paths

import "os"

func userDataDir() (string, error) {
	if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
		return dir, nil
	}
	return "", ErrNoDataDir
}
