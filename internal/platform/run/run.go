// Package run starts the external tools the kit drives.
package run

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
)

// Output is what a finished command printed and how it exited.
type Output struct {
	Stdout []byte
	Stderr []byte
	Code   int
}

// Command runs argv in dir with env added to the caller's environment. It
// returns an error only when the command could not run; a non-zero exit is
// reported in Output.Code.
func Command(ctx context.Context, dir string, env []string, argv ...string) (Output, error) {
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = dir
	cmd.Env = append(cmd.Environ(), env...)
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
