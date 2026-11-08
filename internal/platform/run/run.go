// Package run starts the external tools the kit drives.
package run

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
)

// Output is what a finished command printed and how it exited.
type Output struct {
	Stdout []byte
	Stderr []byte
	Code   int
}

// Cmd is a command to run: where, the variables it adds to the caller's
// environment, its argv, and what it reads on its standard input, if
// anything.
type Cmd struct {
	Dir   string
	Env   []string
	Argv  []string
	Stdin io.Reader
}

// Command runs argv in dir with the caller's environment (see Run).
func Command(ctx context.Context, dir string, argv ...string) (Output, error) {
	return Run(ctx, Cmd{Dir: dir, Argv: argv})
}

// Run runs c. It returns an error only when the command could not run; a
// non-zero exit is reported in Output.Code.
func Run(ctx context.Context, c Cmd) (Output, error) {
	argv := c.Argv
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = c.Dir
	cmd.Env = append(cmd.Environ(), c.Env...)
	cmd.Stdin = c.Stdin
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	out := Output{Stdout: stdout.Bytes(), Stderr: stderr.Bytes()}
	if exit, ok := errors.AsType[*exec.ExitError](err); ok {
		out.Code = exit.ExitCode()
		return out, nil
	}
	if err != nil {
		return out, fmt.Errorf("%s: %w", argv[0], err)
	}
	return out, nil
}
