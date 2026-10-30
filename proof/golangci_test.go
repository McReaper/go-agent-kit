package proof_test

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const golangciTool = "golangci/"

// WHY: lint runs once per system: a file built for one system only is
// invisible to a run for any other.
var targetGOOS = []string{"linux", "windows"}

type golangciReport struct {
	Issues []golangciIssue `json:"Issues"`
	Report struct {
		Linters []struct {
			Name    string `json:"Name"`
			Enabled bool   `json:"Enabled"`
		} `json:"Linters"`
	} `json:"Report"`
}

type golangciIssue struct {
	FromLinter string `json:"FromLinter"`
	Text       string `json:"Text"`
	Pos        struct {
		Filename string `json:"Filename"`
		Line     int    `json:"Line"`
	} `json:"Pos"`
}

func runGolangci(t *testing.T, dir, goos string) golangciReport {
	t.Helper()
	config, err := filepath.Abs("../lint/golangci.yml")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(t.Context(), "golangci-lint", "run", "--config", config, "--path-mode", "abs",
		"--output.json.path", "stdout", "--output.text.path", "stderr", "--show-stats=false", "./...")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOOS="+goos, "GOFLAGS=-mod=readonly")
	out, err := cmd.Output()
	if err != nil && exitCode(err) != 1 {
		t.Fatalf("golangci-lint (GOOS=%s): %v\n%s", goos, err, stderrOf(err))
	}
	var r golangciReport
	if err := json.Unmarshal(out, &r); err != nil {
		t.Fatalf("golangci-lint (GOOS=%s): %v\n%s", goos, err, out)
	}
	return r
}

func stderrOf(err error) string {
	if exit, ok := errors.AsType[*exec.ExitError](err); ok {
		return string(exit.Stderr)
	}
	return ""
}

func exitCode(err error) int {
	if exit, ok := errors.AsType[*exec.ExitError](err); ok {
		return exit.ExitCode()
	}
	return -1
}

func golangciFindings(t *testing.T, dir string) []finding {
	t.Helper()
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	var all []finding
	for _, goos := range targetGOOS {
		for _, is := range runGolangci(t, dir, goos).Issues {
			rel, err := filepath.Rel(abs, is.Pos.Filename)
			if err != nil {
				t.Fatal(err)
			}
			f := finding{tool: golangciTool + is.FromLinter, file: filepath.ToSlash(rel), line: is.Pos.Line, text: is.Text}
			if !slices.Contains(all, f) {
				all = append(all, f)
			}
		}
	}
	return all
}

func TestGolangciFindsNothingInTheCleanTrial(t *testing.T) {
	check(t, nil, golangciTool, golangciFindings(t, cleanDir))
}

func TestGolangciReportsEveryCase(t *testing.T) {
	proveCases(t, golangciTool, golangciFindings)
}

func TestEveryEnabledLinterHasACase(t *testing.T) {
	proven := map[string]bool{}
	for _, c := range loadCases(t) {
		for _, w := range c.wants {
			proven[strings.TrimPrefix(w.tool, golangciTool)] = true
		}
	}
	for _, l := range runGolangci(t, cleanDir, "linux").Report.Linters {
		if l.Enabled && !proven[l.Name] && l.Name != "typecheck" {
			t.Errorf("linter %s is enabled but no case under trial/cases proves it fires", l.Name)
		}
	}
}
