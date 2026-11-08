// Package gitleaks runs gitleaks on a git repository with Rumblestrip's config
// and reads its report.
package gitleaks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/McReaper/rumblestrip-go/internal/platform/run"
	"github.com/McReaper/rumblestrip-go/lint"
)

// Leak is one secret gitleaks found.
type Leak struct {
	Rule        string `json:"RuleID"`
	Description string `json:"Description"`
	File        string `json:"File"`
	Line        int    `json:"StartLine"`
}

// Scan runs gitleaks git on the repository at dir, with scope's options
// (--log-opts=<range>, --pre-commit, --staged), and returns what it found.
//
// WHY: Rumblestrip's config goes with every scan; without --config gitleaks
// reads the repository's own .gitleaks.toml, which could allow a secret.
func Scan(ctx context.Context, dir string, scope ...string) (leaks []Leak, err error) {
	config, err := writeConfig()
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, os.Remove(config)) }()
	argv := append([]string{"gitleaks", "git", "--no-banner", "--redact", "--config", config, "-f", "json", "-r", "-"}, scope...)
	out, err := run.Command(ctx, dir, append(argv, ".")...)
	if err != nil {
		return nil, err
	}
	if out.Code != 0 && out.Code != 1 {
		return nil, fmt.Errorf("gitleaks: exit %d: %s", out.Code, strings.TrimSpace(string(out.Stderr)))
	}
	if err := json.Unmarshal(out.Stdout, &leaks); err != nil {
		return nil, fmt.Errorf("decode gitleaks' report: %w", err)
	}
	return leaks, nil
}

func writeConfig() (string, error) {
	f, err := os.CreateTemp("", "rumblestrip-gitleaks-*.toml")
	if err != nil {
		return "", fmt.Errorf("write the gitleaks config: %w", err)
	}
	_, werr := f.Write(lint.Gitleaks)
	if err := errors.Join(werr, f.Close()); err != nil {
		return "", fmt.Errorf("write the gitleaks config: %w", err)
	}
	return f.Name(), nil
}
