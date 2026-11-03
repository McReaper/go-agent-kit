package proof_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// WHY: a shell function shadows gh, so the job's script runs offline and
// with no executable file to write. Its answers: spec/a merged into main
// and #7 is stacked on it; feat/c merged into feat/b and #9 is stacked on
// it; feat/d was pushed to after it merged; main never goes.
const fakeGH = `gh() {
  echo "$*" >> "$LOG"
  case "$*" in
  *"pulls?state=closed"*) printf '%s\n' "spec/a main aaa" "feat/c feat/b ccc" "feat/d main ddd" "main other mmm" ;;
  *"git/ref/heads/spec/a"*) echo aaa ;;
  *"git/ref/heads/feat/c"*) echo ccc ;;
  *"git/ref/heads/feat/d"*) echo moved ;;
  *"git/ref/heads/main"*) echo mmm ;;
  *"base=spec/a"*) echo 7 ;;
  *"base=feat/c"*) echo 9 ;;
  esac
}
`

// WHY: GitHub's own deletion skips a branch an App's token merged, and a
// pull request stacked on a branch moves to the default branch only once
// that branch is gone; deleting the branch first closes it instead.
func TestTheBranchesJobMovesAStackedPullRequestBeforeDeletingItsBase(t *testing.T) {
	dir := t.TempDir()
	fake, script, log := filepath.Join(dir, "gh.sh"), filepath.Join(dir, "branches.sh"), filepath.Join(dir, "log")
	for path, text := range map[string]string{fake: fakeGH, script: runBlock(t, "../.github/workflows/kit-branches.yml")} {
		if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.CommandContext(t.Context(), "sh", "-c", `. "$1" && . "$2"`, "sh", fake, script)
	cmd.Env = append(os.Environ(), "LOG="+log, "REPO=o/r", "DEFAULT=main", "GH_TOKEN=x")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("the job's script: %v\n%s", err, out)
	}
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	var writes []string
	for line := range strings.Lines(string(data)) {
		if strings.Contains(line, " -X ") || strings.HasPrefix(line, "pr merge ") {
			writes = append(writes, strings.TrimSpace(line))
		}
	}
	want := []string{
		"api -X PATCH repos/o/r/pulls/7 -f base=main --silent",
		"pr merge 7 --repo o/r --auto --squash",
		"api -X DELETE repos/o/r/git/refs/heads/spec/a --silent",
		"api -X PATCH repos/o/r/pulls/9 -f base=feat/b --silent",
		"api -X DELETE repos/o/r/git/refs/heads/feat/c --silent",
	}
	if diff := cmp.Diff(want, writes); diff != "" {
		t.Errorf("the job's writes (-want +got):\n%s", diff)
	}
}

func runBlock(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	_, rest, ok := strings.Cut(string(data), "run: |\n")
	if !ok {
		t.Fatalf("%s has no run block", path)
	}
	var b strings.Builder
	for line := range strings.Lines(rest) {
		body, ok := strings.CutPrefix(line, "          ")
		if !ok && strings.TrimSpace(line) != "" {
			break
		}
		b.WriteString(body)
	}
	return b.String()
}
