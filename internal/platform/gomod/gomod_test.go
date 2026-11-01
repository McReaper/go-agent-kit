package gomod_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/McReaper/go-agent-kit/internal/platform/gomod"
)

// WHY: the kit depends on go-cmp, so the module cache holds it; with the
// proxy off, the test reads the cache and needs no network.
func TestDownloadReturnsTheModulesDirectoryFromTheCache(t *testing.T) {
	t.Setenv("GOPROXY", "off")
	dir, err := gomod.Download(t.Context(), "github.com/google/go-cmp", "v0.7.0")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "cmp", "compare.go")); err != nil || !strings.Contains(filepath.ToSlash(dir), "go-cmp@v0.7.0") {
		t.Errorf("dir = %s (%v), want go-cmp v0.7.0's sources", dir, err)
	}
}

func TestDownloadSaysWhichModuleItCouldntFetch(t *testing.T) {
	t.Setenv("GOPROXY", "off")
	_, err := gomod.Download(t.Context(), "example.com/nowhere", "v1.0.0")
	if err == nil || !strings.Contains(err.Error(), "example.com/nowhere@v1.0.0") {
		t.Errorf("err = %v, want the module named", err)
	}
}
