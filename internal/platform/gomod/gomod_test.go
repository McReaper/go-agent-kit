package gomod_test

import (
	"strings"
	"testing"

	"github.com/McReaper/go-agent-kit/internal/platform/gomod"
)

// WHY: the kit depends on go-cmp, so the module cache holds it; with the
// proxy off, the test reads the cache and needs no network.
func TestCommitReturnsTheCommitAVersionNames(t *testing.T) {
	t.Setenv("GOPROXY", "off")
	got, err := gomod.Commit(t.Context(), "github.com/google/go-cmp", "v0.7.0")
	if err != nil || got != "9b12f366a942ebc7254abc7f32ca05068b455fb7" {
		t.Errorf("Commit = %q, %v; want go-cmp v0.7.0's commit", got, err)
	}
}

func TestCommitSaysWhichModuleItCouldntFetch(t *testing.T) {
	t.Setenv("GOPROXY", "off")
	_, err := gomod.Commit(t.Context(), "example.com/nowhere", "v1.0.0")
	if err == nil || !strings.Contains(err.Error(), "example.com/nowhere@v1.0.0") {
		t.Errorf("err = %v, want the module named", err)
	}
}
