// Package psscriptanalyzer runs PSScriptAnalyzer on PowerShell scripts.
package psscriptanalyzer

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/McReaper/go-agent-kit/internal/platform/run"
)

// Version is the PSScriptAnalyzer the kit's proofs ran with; tools.txt
// names the same.
const Version = "1.25.0"

// Finding is one rule a script breaks.
type Finding struct {
	File    string
	Line    int
	Rule    string
	Message string
}

// Analyze runs PSScriptAnalyzer on files, paths relative to dir, at
// warning severity and above. It runs in pwsh, or in Windows PowerShell
// where pwsh is absent.
func Analyze(ctx context.Context, dir string, files []string) ([]Finding, error) {
	shell := "pwsh"
	if _, err := exec.LookPath(shell); err != nil {
		shell = "powershell"
	}
	out, err := run.Command(ctx, dir, nil, shell, "-NoProfile", "-NonInteractive", "-Command", script(files))
	if err != nil {
		return nil, err
	}
	if out.Code != 0 {
		return nil, fmt.Errorf("PSScriptAnalyzer %s: exit %d: %s", Version, out.Code, strings.TrimSpace(string(out.Stderr)))
	}
	return parse(string(out.Stdout))
}

// WHY: the paths come from git and go into single-quoted PowerShell
// strings, where only a quote needs escaping, doubled; the tab between
// fields needs double quotes, where PowerShell reads `t.
func script(files []string) string {
	quoted := make([]string, 0, len(files))
	for _, f := range files {
		quoted = append(quoted, "'"+strings.ReplaceAll(f, "'", "''")+"'")
	}
	return "$ErrorActionPreference = 'Stop'; Import-Module PSScriptAnalyzer -RequiredVersion " + Version + "; " +
		"foreach ($f in @(" + strings.Join(quoted, ",") + ")) { " +
		"Invoke-ScriptAnalyzer -Path $f -Severity Warning,Error | ForEach-Object { " +
		"\"{0}`t{1}`t{2}`t{3}\" -f $f, $_.Line, $_.RuleName, ($_.Message -replace '\\s+', ' ') } }"
}

func parse(out string) ([]Finding, error) {
	var fs []Finding
	for line := range strings.Lines(out) {
		fields := strings.SplitN(strings.TrimRight(line, "\r\n"), "\t", 4)
		if len(fields) != 4 {
			continue
		}
		n, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, fmt.Errorf("PSScriptAnalyzer: line %q: %w", line, err)
		}
		fs = append(fs, Finding{File: fields[0], Line: n, Rule: fields[2], Message: fields[3]})
	}
	return fs, nil
}
