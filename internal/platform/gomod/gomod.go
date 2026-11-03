// Package gomod reads modules through the go command.
package gomod

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/McReaper/go-agent-kit/internal/platform/run"
)

// Commit returns the commit module's version names, from the module proxy's
// record of where the version came from.
func Commit(ctx context.Context, module, version string) (string, error) {
	m, err := download(ctx, module, version)
	if err == nil && len(m.Origin.Hash) != 40 {
		err = fmt.Errorf("go mod download %s@%s: no commit", module, version)
	}
	return m.Origin.Hash, err
}

type downloaded struct {
	Origin struct {
		Hash string `json:"Hash"`
	} `json:"Origin"`
}

func download(ctx context.Context, module, version string) (downloaded, error) {
	o, err := run.Command(ctx, ".", "go", "mod", "download", "-json", module+"@"+version)
	if err != nil {
		return downloaded{}, err
	}
	if o.Code != 0 {
		return downloaded{}, fmt.Errorf("go mod download %s@%s: %s", module, version, strings.TrimSpace(string(o.Stderr)+string(o.Stdout)))
	}
	var m downloaded
	if err := json.Unmarshal(o.Stdout, &m); err != nil {
		return downloaded{}, fmt.Errorf("go mod download %s@%s: %w", module, version, err)
	}
	return m, nil
}
