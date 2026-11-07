package proof_test

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// WHY: a pull request of a stack takes no auto-merge, and GitHub rebases
// the next one onto the default branch only once the one below merged: the
// kit's merge job merges through the asynchronous merge, and says why when
// the ruleset refuses.
func TestTheMergeJobMergesThroughTheAsynchronousMergeAndSaysWhyGitHubRefused(t *testing.T) {
	script := stepRun(t, "../.github/workflows/kit-check.yml", "Merge")
	for _, c := range []struct {
		name, state, result string
		code                int
		writes              []string
		says                string
	}{
		{"green", "open false h1", "merged", 0, []string{"api -H X-GitHub-Api-Version: 2026-03-10 -X PUT repos/o/r/pulls/7/merge-async -f merge_method=squash -f sha=h1 --jq .details.uuid // empty"}, "merged"},
		{"refused", "open false h1", "failed Repository rule violations found", 1, []string{"api -H X-GitHub-Api-Version: 2026-03-10 -X PUT repos/o/r/pulls/7/merge-async -f merge_method=squash -f sha=h1 --jq .details.uuid // empty"}, "GitHub refused the merge: Repository rule violations found"},
		{"merged by an earlier run", "closed true h1", "", 0, nil, "already merged"},
		{"pushed to since", "open false h2", "", 0, nil, "its own run merges it"},
	} {
		t.Run(c.name, func(t *testing.T) {
			fake := `gh() {
  echo "$*" >> "$LOG"
  case "$*" in
  *merge-async/u1*) echo "$RESULT" ;;
  *merge-async*) echo u1 ;;
  *pulls/7*) echo "$STATE" ;;
  esac
}
sleep() { :; }
`
			out, code, log := runStep(t, fake, script, "PR=7", "HEAD=h1", "STATE="+c.state, "RESULT="+c.result)
			writes := gestures(log)
			if code != c.code || !strings.Contains(out, c.says) {
				t.Errorf("exit %d, want %d; output lacks %q:\n%s", code, c.code, c.says, out)
			}
			if strings.Join(writes, "\n") != strings.Join(c.writes, "\n") {
				t.Errorf("writes %q, want %q", writes, c.writes)
			}
		})
	}
}

// WHY: GitHub moves a pull request onto the default branch once the one
// below it merged only when both are a stack; outside one it stays on a
// merged branch and never merges. The check says so, with the command.
func TestAPullRequestOnAnotherBranchIsInAStackOrTheCheckSaysHow(t *testing.T) {
	script := stepRun(t, "../.github/workflows/kit-check.yml", "Is a pull request on another branch than the default one in a stack?")
	fake := `gh() { echo "$*" >> "$LOG"; echo "$STACK"; }
sleep() { :; }
`
	out, code, _ := runStep(t, fake, script, "PR=9", "BASE_PR=spec/x", "STACK=3")
	if code != 0 {
		t.Errorf("in a stack: exit %d\n%s", code, out)
	}
	out, code, _ = runStep(t, fake, script, "PR=9", "BASE_PR=spec/x", "STACK=")
	if code != 1 || !strings.Contains(out, "gh api -X POST repos/o/r/stacks") || !strings.Contains(out, "pull_requests[]=9") {
		t.Errorf("in no stack: exit %d\n%s", code, out)
	}
}

// WHY: the run that failed for want of the owner's approval stays on the
// commit, and GitHub holds the merge on it even once another run passes:
// the approval re-runs that run, once it finished, and no other.
func TestAnApprovalReRunsTheMergeCheckThatFailedOnTheApprovedCommit(t *testing.T) {
	script := stepRun(t, "../.github/workflows/kit-approval.yml", "Re-run the merge check that failed on the approved commit")
	fake := `gh() {
  echo "$*" >> "$LOG"
  case "$*" in
  *actions/runs*)
    if [ -e "$LOG.listed" ]; then
      printf '%s\n' "11 completed failure" "12 completed failure" "13 completed success"
    else
      : > "$LOG.listed"
      printf '%s\n' "11 in_progress " "12 completed failure" "13 completed success"
    fi ;;
  esac
}
`
	out, code, log := runStep(t, fake, script, "HEAD=h1", "CALLER=o/r/.github/workflows/kit.yml@refs/pull/6/merge")
	if code != 0 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	want := []string{"run watch 11 --repo o/r", "run rerun 11 --repo o/r --failed", "run rerun 12 --repo o/r --failed"}
	if got := gestures(log); strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("gestures %q, want %q", got, want)
	}
	if !strings.Contains(log, `select(.path == ".github/workflows/kit.yml")`) {
		t.Errorf("the runs listed are not the caller's only:\n%s", log)
	}
}

func runStep(t *testing.T, fake, script string, env ...string) (string, int, string) {
	t.Helper()
	dir := t.TempDir()
	fakePath, scriptPath, log := filepath.Join(dir, "gh.sh"), filepath.Join(dir, "step.sh"), filepath.Join(dir, "log")
	for path, text := range map[string]string{fakePath: fake, scriptPath: script} {
		if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.CommandContext(t.Context(), "sh", "-c", `. "$1" && . "$2"`, "sh", fakePath, scriptPath)
	cmd.Env = append(append(os.Environ(), "LOG="+log, "REPO=o/r", "GH_TOKEN=x"), env...)
	out, err := cmd.CombinedOutput()
	code := 0
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		code = exit.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(log)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		t.Fatal(err)
	}
	return string(out), code, string(data)
}

func gestures(log string) []string {
	var out []string
	for line := range strings.Lines(log) {
		if strings.Contains(line, " -X PUT ") || strings.HasPrefix(line, "run ") {
			out = append(out, strings.TrimSpace(line))
		}
	}
	return out
}

func stepRun(t *testing.T, path, step string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	_, rest, ok := strings.Cut(string(data), "- name: "+step+"\n")
	if !ok {
		t.Fatalf("%s has no step %q", path, step)
	}
	head, rest, ok := strings.Cut(rest, "run: |\n")
	if !ok || strings.Contains(head, "- name: ") {
		t.Fatalf("step %q of %s has no run block", step, path)
	}
	indent := head[strings.LastIndex(head, "\n")+1:] + "  "
	var b strings.Builder
	for line := range strings.Lines(rest) {
		body, ok := strings.CutPrefix(line, indent)
		if !ok && strings.TrimSpace(line) != "" {
			break
		}
		b.WriteString(body)
	}
	return b.String()
}
