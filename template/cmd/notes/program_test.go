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

// WHY: the built program and the data directory it runs on, as a user's
// would be.
type program struct {
	bin     string
	dataDir string
}

func (p program) command(ctx context.Context, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, p.bin, args...)
	cmd.Env = append(os.Environ(), "NOTES_DATA_DIR="+p.dataDir)
	return cmd
}

func (p program) run(t *testing.T, args ...string) result {
	t.Helper()
	var stdout, stderr bytes.Buffer
	cmd := p.command(t.Context(), args...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	var exit *exec.ExitError
	if err != nil && !errors.As(err, &exit) {
		t.Fatalf("notes %v: %v", args, err)
	}
	return result{code: cmd.ProcessState.ExitCode(), stdout: stdout.String(), stderr: stderr.String()}
}

func (p program) connect(t *testing.T) *mcp.ClientSession {
	t.Helper()
	c := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	// WHY: t.Context ends before the cleanups run, and would kill the server
	// before Close asks it to stop.
	server := p.command(context.WithoutCancel(t.Context()), "mcp")
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
