package proof_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// WHY: the workflow writes, for each commit a code owner approved, its tree
// as `<blob> <path>` lines from GitHub's API; this writes the same from git.
func approveHead(t *testing.T, dir, trees string) {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), "git", "ls-tree", "-r", "--format=%(objectname) %(path)", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git ls-tree: %v", err)
	}
	if err := os.MkdirAll(trees, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(trees, "approved"), out, 0o600); err != nil {
		t.Fatal(err)
	}
}

func commitFile(t *testing.T, dir, name, text string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(name)), []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	gitIn(t, dir, "add", "-A")
	gitIn(t, dir, "commit", "-q", "-m", "after the approval: "+name)
}

// WHY: approving the last commit only made the owner approve the same
// protected change after each push, and pushed the agent to stop asking.
// The approval holds the protected files' content: the code goes on
// without asking again, and touching a protected file reopens it.
func TestTheMergeCheckKeepsAnApprovalWhileTheProtectedFilesStayAsApproved(t *testing.T) {
	dir := changeOf(t, caseNamed(t, "guard-protected"))
	trees := t.TempDir()
	approveHead(t, dir, trees)
	commitFile(t, dir, "README.md", "notes, with more said about them\n")
	if out, code := guardRun(t, dir, "--ci", "--approved-trees", trees); code != 0 {
		t.Errorf("the code moved on, the protected files as approved: exit %d, want 0\n%s", code, out)
	}
	commitFile(t, dir, "internal/notes/internal/add/add_test.go", "package add_test\n")
	if out, code := guardRun(t, dir, "--ci", "--approved-trees", trees); code != 1 {
		t.Errorf("a protected file touched after the approval: exit %d, want 1\n%s", code, out)
	}
}
