package proof_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"testing"
)

const guardTool = "guard/"

type guardRow struct {
	Rule    string `json:"rule"`
	File    string `json:"file"`
	Line    int    `json:"line"`
	Message string `json:"message"`
}

// WHY: the guard reads a change, not a tree: each case becomes a commit on
// top of the clean trial, the way a pull request sits on main.
func changeOf(t *testing.T, c trialCase) string {
	t.Helper()
	dir := overlay(t)
	gitIn(t, dir, "init", "-q")
	gitIn(t, dir, "add", "-A")
	gitIn(t, dir, "commit", "-q", "-m", "clean trial")
	gitIn(t, dir, "branch", "base")
	apply(t, c, dir)
	if !c.uncommitted {
		gitIn(t, dir, "add", "-A")
		gitIn(t, dir, "commit", "-q", "--allow-empty", "-m", c.name)
	}
	return dir
}

func gitIn(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), "git", append([]string{
		"-c", "user.name=proof", "-c", "user.email=proof@example.com",
		"-c", "core.autocrlf=false", "-c", "commit.gpgsign=false",
	}, args...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func guardFindings(t *testing.T, dir string) []finding {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), kitCommand(t, "kit"), "guard", "--base", "base", "--json")
	cmd.Dir = dir
	cmd.Env = os.Environ()
	out, err := cmd.Output()
	if err != nil && exitCode(err) != 1 {
		t.Fatalf("kit guard: %v\n%s", err, stderrOf(err))
	}
	var rows []guardRow
	if err := json.Unmarshal(out, &rows); err != nil {
		t.Fatalf("kit guard: %v\n%s", err, out)
	}
	findings := make([]finding, 0, len(rows))
	for _, r := range rows {
		findings = append(findings, finding{tool: guardTool + r.Rule, file: r.File, line: r.Line, text: r.Message})
	}
	return findings
}

func TestGuardReportsEveryCase(t *testing.T) {
	for _, c := range loadCases(t) {
		if !c.uses(guardTool) {
			continue
		}
		if !c.isolated {
			t.Errorf("case %s: a guard case is one change: mark it !isolated", c.name)
			continue
		}
		t.Run(c.name, func(t *testing.T) {
			check(t, []trialCase{c}, guardTool, guardFindings(t, changeOf(t, c)))
		})
	}
}
