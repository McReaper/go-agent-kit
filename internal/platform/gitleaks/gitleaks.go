// Package gitleaks runs gitleaks on a git repository and reads its report.
package gitleaks

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/McReaper/go-agent-kit/internal/platform/run"
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
func Scan(ctx context.Context, dir string, scope ...string) ([]Leak, error) {
	argv := append([]string{"gitleaks", "git", "--no-banner", "--redact", "-f", "json", "-r", "-"}, scope...)
	out, err := run.Command(ctx, dir, nil, append(argv, ".")...)
	if err != nil {
		return nil, err
	}
	if out.Code != 0 && out.Code != 1 {
		return nil, fmt.Errorf("gitleaks: exit %d: %s", out.Code, strings.TrimSpace(string(out.Stderr)))
	}
	var leaks []Leak
	if err := json.Unmarshal(out.Stdout, &leaks); err != nil {
		return nil, fmt.Errorf("decode gitleaks' report: %w", err)
	}
	return leaks, nil
}
