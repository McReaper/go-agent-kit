package proof_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var (
	bin      string
	buildErr error
)

// WHY: the kit's commands are built once for the whole run, not per test;
// a failed build fails every test that needs them, with its output.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "kit-proof-bin")
	bin, buildErr = dir, err
	if err == nil {
		buildErr = build("kitvet", "kit")
	}
	m.Run()
	if err := os.RemoveAll(dir); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func build(cmds ...string) error {
	for _, c := range cmds {
		out, err := exec.CommandContext(context.Background(), "go", "build", "-o", filepath.Join(bin, c+".exe"), "../cmd/"+c).CombinedOutput()
		if err != nil {
			return fmt.Errorf("build %s: %w\n%s", c, err, out)
		}
	}
	return nil
}

func kitCommand(t *testing.T, name string) string {
	t.Helper()
	if buildErr != nil {
		t.Fatal(buildErr)
	}
	return filepath.Join(bin, name+".exe")
}
