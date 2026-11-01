// Package gomod reads modules through the go command.
package gomod

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/McReaper/go-agent-kit/internal/platform/run"
)

// Download returns the directory of module at version in the module cache,
// fetching it first if the cache doesn't hold it.
func Download(ctx context.Context, module, version string) (string, error) {
	o, err := run.Command(ctx, ".", nil, "go", "mod", "download", "-json", module+"@"+version)
	if err != nil {
		return "", err
	}
	if o.Code != 0 {
		return "", fmt.Errorf("go mod download %s@%s: %s", module, version, strings.TrimSpace(string(o.Stderr)+string(o.Stdout)))
	}
	var m struct {
		Dir string `json:"Dir"`
	}
	if err := json.Unmarshal(o.Stdout, &m); err != nil || m.Dir == "" {
		return "", fmt.Errorf("go mod download %s@%s: no directory in %q", module, version, o.Stdout)
	}
	return m.Dir, nil
}
