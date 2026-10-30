package proof

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

// targetGOOS are the systems the trial builds for; lint runs once per
// system, since a file built for one only is invisible to the others.
var targetGOOS = []string{"linux", "windows"}

type golangciReport struct {
	Issues []struct {
		FromLinter string
		Text       string
		Pos        struct {
			Filename string
			Line     int
		}
	}
	Report struct {
		Linters []struct {
			Name    string
			Enabled bool
		}
	}
}

func runGolangci(t *testing.T, dir, goos string) golangciReport {
	t.Helper()
	config, err := filepath.Abs("../lint/golangci.yml")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("golangci-lint", "run", "--config", config, "--path-mode", "abs",
		"--output.json.path", "stdout", "--output.text.path", "stderr", "--show-stats=false", "./...")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOOS="+goos, "GOFLAGS=-mod=readonly")
	out, err := cmd.Output()
	var exit *exec.ExitError
	if err != nil && !(errors.As(err, &exit) && exit.ExitCode() == 1) {
		t.Fatalf("golangci-lint (GOOS=%s): %v\n%s", goos, err, stderrOf(err))
	}
	var r golangciReport
	if err := json.Unmarshal(out, &r); err != nil {
		t.Fatalf("golangci-lint (GOOS=%s): %v\n%s", goos, err, out)
	}
	return r
}

func stderrOf(err error) string {
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		return string(exit.Stderr)
	}
	return ""
}

// golangciFindings lints dir for every target system and returns the union
// of what it reports.
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
	var together []trialCase
	for _, c := range loadCases(t) {
		switch {
		case !c.uses(golangciTool):
		case c.isolated:
			t.Run(c.name, func(t *testing.T) {
				check(t, []trialCase{c}, golangciTool, golangciFindings(t, overlay(t, c)))
			})
		default:
			together = append(together, c)
		}
	}
	check(t, together, golangciTool, golangciFindings(t, overlay(t, together...)))
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
