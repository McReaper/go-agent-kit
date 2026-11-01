package main_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type result struct {
	code   int
	stdout string
	stderr string
}

// WHY: these tests run the program people run, built from this package, not
// its code called in memory.
func build(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "notes")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	if out, err := exec.CommandContext(t.Context(), "go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	return bin
}

func command(ctx context.Context, bin, dataDir string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Env = append(os.Environ(), "NOTES_DATA_DIR="+dataDir)
	return cmd
}

func run(t *testing.T, bin, dataDir string, args ...string) result {
	t.Helper()
	var stdout, stderr bytes.Buffer
	cmd := command(t.Context(), bin, dataDir, args...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	var exit *exec.ExitError
	if err != nil && !errors.As(err, &exit) {
		t.Fatalf("notes %v: %v", args, err)
	}
	return result{code: cmd.ProcessState.ExitCode(), stdout: stdout.String(), stderr: stderr.String()}
}

func connect(t *testing.T, bin, dataDir string) *mcp.ClientSession {
	t.Helper()
	c := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	// WHY: t.Context ends before the cleanups run, and would kill the server
	// before Close asks it to stop.
	server := command(context.WithoutCancel(t.Context()), bin, dataDir, "mcp")
	s, err := c.Connect(t.Context(), &mcp.CommandTransport{Command: server}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Error(err)
		}
	})
	return s
}
