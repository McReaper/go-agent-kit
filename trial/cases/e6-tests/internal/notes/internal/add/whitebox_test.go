package add

import (
	"os"
	"testing"
)

func scratch(t *testing.T) string {
	dir, err := os.MkdirTemp("", "notes")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestScratchIsADirectory(t *testing.T) {
	fi, err := os.Stat(scratch(t))
	if err != nil || !fi.IsDir() {
		t.Fatalf("Stat = %v, %v", fi, err)
	}
}
