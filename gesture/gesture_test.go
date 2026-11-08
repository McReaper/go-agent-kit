package gesture_test

import (
	"path/filepath"
	"testing"

	"github.com/McReaper/rumblestrip-go/gesture"
)

func TestTheGesturesPastTheKitsJudgementAreRefused(t *testing.T) {
	for _, line := range []string{
		"gh pr merge 12 --squash",
		"git push --force origin feat/x",
		"git push -f",
		"git push origin +feat/x",
		"git -C . push origin main",
		"git push origin HEAD:main",
		"git push --force-with-lease origin main",
		"go test ./... && git commit --no-verify -m wip",
		"git -c core.hooksPath=/dev/null commit -n -m wip",
		"gh api repos/o/r/rulesets/1 --method PUT --input x.json",
	} {
		if gesture.Check(gesture.Tool{Name: "Bash", Command: line}) == "" {
			t.Errorf("%q passes", line)
		}
	}
}

func TestTheAgentsOwnWorkFlowsFreely(t *testing.T) {
	for _, line := range []string{
		"git push --force-with-lease origin feat/x",
		"git push -u origin spec/export",
		"git commit -m 'feat: export'",
		"gh pr create --base spec/export --fill",
		"rumblestrip check",
		"git log --oneline main",
	} {
		if r := gesture.Check(gesture.Tool{Name: "Bash", Command: line}); r != "" {
			t.Errorf("%q refused: %s", line, r)
		}
	}
}

func TestAnEditOfWhatRunsTheChecksIsRefusedButNotOfTheCodeOrItsInstructions(t *testing.T) {
	dir := t.TempDir()
	for file, refused := range map[string]bool{
		".github/workflows/rumblestrip.yml": true,
		".claude/settings.json":             true,
		".rumblestrip/sync":                 true,
		"AGENTS.md":                         false,
		"internal/notes/notes.go":           false,
		"features/notes.feature":            false,
	} {
		got := gesture.Check(gesture.Tool{Name: "Edit", FilePath: filepath.Join(dir, filepath.FromSlash(file)), Dir: dir}) != ""
		if got != refused {
			t.Errorf("an edit of %s refused = %v, want %v", file, got, refused)
		}
	}
}
