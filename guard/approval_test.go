package guard_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/McReaper/go-agent-kit/guard"
	"github.com/McReaper/go-agent-kit/internal/platform/git"
	"github.com/google/go-cmp/cmp"
)

type repo struct {
	t   *testing.T
	dir string
}

func newRepo(t *testing.T) repo {
	t.Helper()
	r := repo{t, t.TempDir()}
	r.git("init", "-q")
	r.git("config", "user.email", "test@example.com")
	r.git("config", "user.name", "test")
	r.git("config", "commit.gpgsign", "false")
	return r
}

func (r repo) git(args ...string) string {
	r.t.Helper()
	out, err := git.Run(r.t.Context(), r.dir, args...)
	if err != nil {
		r.t.Fatal(err)
	}
	return strings.TrimSpace(out)
}

func (r repo) commit(files map[string]string, deleted ...string) {
	r.t.Helper()
	for name, text := range files {
		if err := os.WriteFile(filepath.Join(r.dir, name), []byte(text), 0o600); err != nil {
			r.t.Fatal(err)
		}
	}
	for _, name := range deleted {
		r.git("rm", "-q", name)
	}
	r.git("add", "-A")
	r.git("commit", "-q", "-m", "change")
}

// WHY: the workflow writes each approved commit's tree as `<blob> <path>`
// lines, from GitHub's API; this builds the same from git.
func (r repo) tree() guard.Tree {
	r.t.Helper()
	t, err := guard.ParseTree(r.git("ls-tree", "-r", "--format=%(objectname) %(path)", "HEAD"))
	if err != nil {
		r.t.Fatal(err)
	}
	return t
}

func (r repo) covered(fs []guard.Finding, a guard.Approval) []string {
	r.t.Helper()
	got, err := guard.Covered(r.t.Context(), r.dir, fs, a)
	if err != nil {
		r.t.Fatal(err)
	}
	return rules(got)
}

var changedTest = []guard.Finding{{Rule: "protected-test", File: "x_test.go"}}

func TestAnApprovedTestStaysApprovedUntilItsContentChanges(t *testing.T) {
	r := newRepo(t)
	r.commit(map[string]string{"x_test.go": "package x // v1\n"})
	r.commit(map[string]string{"x_test.go": "package x // v2, the change the owner approves\n"})
	approved := guard.Approval{Trees: []guard.Tree{r.tree()}}
	r.commit(map[string]string{"x.go": "package x // the code goes on\n"})
	if diff := cmp.Diff([]string{"protected-test"}, r.covered(changedTest, approved)); diff != "" {
		t.Errorf("the test as approved, the code moved on (-want +got):\n%s", diff)
	}
	r.commit(map[string]string{"x_test.go": "package x // v3, touched after the approval\n"})
	if got := r.covered(changedTest, approved); len(got) != 0 {
		t.Errorf("the test touched after the approval: covered %v, want none", got)
	}
}

func TestAnApprovedDeletionStaysApprovedWhileTheFileStaysDeleted(t *testing.T) {
	r := newRepo(t)
	r.commit(map[string]string{"x_test.go": "package x\n", "x.go": "package x\n"})
	r.commit(nil, "x_test.go")
	approved := guard.Approval{Trees: []guard.Tree{r.tree()}}
	if diff := cmp.Diff([]string{"protected-test"}, r.covered(changedTest, approved)); diff != "" {
		t.Errorf("deleted as approved (-want +got):\n%s", diff)
	}
	r.commit(map[string]string{"x_test.go": "package x // back again\n"})
	if got := r.covered(changedTest, approved); len(got) != 0 {
		t.Errorf("brought back after an approved deletion: covered %v, want none", got)
	}
}

func TestOnlyAnApprovedHeadCoversASizeAndBrokenRulesNever(t *testing.T) {
	r := newRepo(t)
	r.commit(map[string]string{"x.go": "package x\n"})
	fs := []guard.Finding{{Rule: "size-lines"}, {Rule: "bypass-skip", File: "x.go"}}
	if got := r.covered(fs, guard.Approval{Trees: []guard.Tree{r.tree()}}); len(got) != 0 {
		t.Errorf("an approved tree covered %v, want none", got)
	}
	if diff := cmp.Diff([]string{"size-lines"}, r.covered(fs, guard.Approval{Head: true})); diff != "" {
		t.Errorf("an approved head (-want +got):\n%s", diff)
	}
}

func TestReadTreesReadsOneTreePerFileAndNoneWithoutADirectory(t *testing.T) {
	dir := t.TempDir()
	for name, text := range map[string]string{"c1": "aaa x_test.go\nbbb dir with space/y.go\n", "c2": "ccc x_test.go\n"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	got, err := guard.ReadTrees(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []guard.Tree{{"x_test.go": "aaa", "dir with space/y.go": "bbb"}, {"x_test.go": "ccc"}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("ReadTrees (-want +got):\n%s", diff)
	}
	for _, missing := range []string{"", filepath.Join(dir, "nowhere")} {
		if got, err := guard.ReadTrees(missing); err != nil || len(got) != 0 {
			t.Errorf("ReadTrees(%q) = %v, %v; want none", missing, got, err)
		}
	}
	if _, err := guard.ParseTree("aaa\n"); err == nil {
		t.Error("ParseTree accepted a line without a path")
	}
}
