// Package git runs git commands in a repository.
package git

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// ErrOptionLikeRef is returned for a ref git would read as an option.
var ErrOptionLikeRef = errors.New("a ref may not start with '-'")

// Run runs git with args in dir and returns its standard output.
func Run(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return string(out), nil
}

// CheckRef refuses a ref that git would take for an option.
func CheckRef(ref string) error {
	if strings.HasPrefix(ref, "-") {
		return fmt.Errorf("%q: %w", ref, ErrOptionLikeRef)
	}
	return nil
}
