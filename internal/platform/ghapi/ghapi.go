// Package ghapi calls GitHub's REST API through gh, as whoever gh acts as:
// the owner logged in, or the agent's App inside rumblestrip app exec.
package ghapi

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/McReaper/rumblestrip-go/internal/platform/run"
)

// Request is one call to GitHub's REST API. Version is the API version it
// asks for, the default one when empty; the stacks API needs 2026-03-10.
type Request struct {
	Method  string
	Path    string
	Version string
	Body    []byte
}

// API calls GitHub's REST API.
type API interface {
	Call(ctx context.Context, r Request) ([]byte, error)
}

// GH calls the API through gh. A path may hold gh's {owner} and {repo},
// which it fills from the repository the command runs in.
type GH struct{}

// Call runs gh api with r; a body goes through a temporary file, so no
// command line carries it.
func (GH) Call(ctx context.Context, r Request) (out []byte, err error) {
	argv := []string{"gh", "api", "--method", r.Method, r.Path}
	if r.Version != "" {
		argv = append(argv, "-H", "X-GitHub-Api-Version: "+r.Version)
	}
	if r.Body != nil {
		f, err := os.CreateTemp("", "rumblestrip-api-*.json")
		if err != nil {
			return nil, fmt.Errorf("gh api: %w", err)
		}
		defer func() { err = errors.Join(err, os.Remove(f.Name())) }()
		_, werr := f.Write(r.Body)
		if err := errors.Join(werr, f.Close()); err != nil {
			return nil, fmt.Errorf("gh api: %w", err)
		}
		argv = append(argv, "--input", f.Name())
	}
	res, err := run.Command(ctx, ".", argv...)
	if err != nil {
		return nil, err
	}
	if res.Code != 0 {
		return nil, fmt.Errorf("gh api %s %s: %s", r.Method, r.Path, strings.TrimSpace(string(res.Stderr)+string(res.Stdout)))
	}
	return res.Stdout, nil
}
