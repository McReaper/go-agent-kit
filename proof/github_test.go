package proof_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	kitcheck "github.com/McReaper/go-agent-kit/check"
)

// WHY: CI runs each gate in a job of its own, which the checks list names
// for the owner; a gate kit check runs and CI lists under no job, or under
// another name, never runs on a pull request, or reads differently on the
// owner's page and in the log.
func TestCIRunsEachGateOfKitCheckInAJobOfItsOwn(t *testing.T) {
	data, err := os.ReadFile("../.github/workflows/kit-check.yml")
	if err != nil {
		t.Fatal(err)
	}
	var got []kitcheck.Gate
	for _, m := range regexp.MustCompile(`(?m)^          - \{ id: ([a-z0-9-]+), title: (.+) \}$`).FindAllStringSubmatch(string(data), -1) {
		got = append(got, kitcheck.Gate{ID: m[1], Title: m[2]})
	}
	var want []kitcheck.Gate
	for _, g := range kitcheck.Gates(kitcheck.Options{GOOS: []string{"linux", "windows"}, Title: "ci: the title CI passes"}) {
		want = append(want, kitcheck.Gate{ID: g.ID, Title: g.Title})
	}
	want = append(want, kitcheck.Gate{ID: "mutate", Title: "Mutation"})
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("kit-check.yml's gate jobs (-kit check's +the workflow's):\n%s", diff)
	}
}

// WHY: on GitHub, a finding is read in two places: on its line, in the
// pull request's files, and in the job's summary, where the owner reads
// what the gate wants and whose the finding is to fix. The log keeps the
// agent's line.
func TestKitCheckAnnotatesEachFindingsLineAndWritesTheOwnersSummary(t *testing.T) {
	dir := changeOf(t, caseNamed(t, "m5-only-tests"))
	summary := filepath.Join(t.TempDir(), "summary.md")
	out, code := kitRun(t, dir, "check", "--base", "base", "--only", "sweep,archtest", "--github", "--summary", summary)
	if code != 1 {
		t.Fatalf("exit %d\n%s", code, out)
	}
	annotation := regexp.MustCompile(`(?m)^::error file=internal/[^,]+\.go,line=\d+,title=Sweep::only-tests: .*Oldest is reached by tests only.*$`)
	if !annotation.MatchString(out) || !strings.Contains(out, "FAIL sweep\n") || !strings.Contains(out, "kit check: 1 gate fails: sweep\n") {
		t.Errorf("the log lacks the annotation or the agent's lines:\n%s", out)
	}
	got, err := os.ReadFile(summary)
	if err != nil {
		t.Fatal(err)
	}
	row := regexp.MustCompile("(?m)^\\| `internal/[^`]+\\.go:\\d+` \\| only-tests: .*Oldest is reached by tests only.* \\|$")
	if want := "### ❌ Sweep\n\nIt wants: no code only tests reach, no name exported for its own package, no log message written twice. It found 2 problems, for the agent to fix:\n\n| Where | What |\n| --- | --- |\n"; !strings.HasPrefix(string(got), want) {
		t.Errorf("the summary lacks %q:\n%s", want, got)
	}
	if strings.Contains(string(got), "Architecture") {
		t.Errorf("the summary has a section for a gate that passed:\n%s", got)
	}
	if !row.Match(got) {
		t.Errorf("the summary lacks the finding's row:\n%s", got)
	}
}

// WHY: go test's own lines name a package and a test, not why it failed:
// the finding is the failing check's message, at its line, so the owner
// reads what broke and the annotation lands on it. A line outside the
// module, godog's, stays in the message.
func TestKitCheckSaysWhichTestFailedWhereAndWhy(t *testing.T) {
	dir := changeOf(t, caseNamed(t, "m6-broken-program"))
	out, code := kitRun(t, dir, "check", "--base", "base", "--only", "go-test", "--race=false", "--github")
	for _, want := range []*regexp.Regexp{
		regexp.MustCompile(`(?m)^  cmd/notes/smoke_test\.go:33: TestTheBuiltProgramAnswersEveryToolAndCommandWithItsExample\S*: the program serves no tool`),
		regexp.MustCompile(`(?m)^::error file=cmd/notes/smoke_test\.go,line=33,title=Tests::TestTheBuiltProgramAnswersEveryToolAndCommandWithItsExample`),
	} {
		if code != 1 || !want.MatchString(out) {
			t.Errorf("exit %d; the output lacks %s:\n%s", code, want, out)
		}
	}
	if strings.Contains(out, "  FAIL") || strings.Contains(out, "--- FAIL") {
		t.Errorf("go test's own lines are findings:\n%s", out)
	}
}

// WHY: the title becomes the commit's subject on the default branch and
// the line GitHub lists the pull request by.
func TestTheTitleGateWantsAShortConventionalCommit(t *testing.T) {
	dir := t.TempDir()
	for _, c := range []struct {
		title string
		code  int
		says  string
	}{
		{"feat(notes): a title shorter than 3 characters is refused", 0, "ok   title\n"},
		{"spec(notes): a title shorter than 3 characters is refused", 0, "ok   title\n"},
		{"fix!: the list starts at the newest note", 0, "ok   title\n"},
		{"Refuse short titles", 1, "title: the title is no conventional commit"},
		{"feat(notes): " + strings.Repeat("a", 60), 1, "title: the title has 73 characters, the limit is 72"},
	} {
		out, code := kitRun(t, dir, "check", "--only", "title", "--title", c.title)
		if code != c.code || !strings.Contains(out, c.says) {
			t.Errorf("%q: exit %d, want %d and %q\n%s", c.title, code, c.code, c.says, out)
		}
	}
}

func TestKitCheckNamesTheGatesWhenAskedForOneItLacks(t *testing.T) {
	cmd := exec.CommandContext(t.Context(), kitCommand(t, "kit"), "check", "--only", "lint")
	cmd.Dir = t.TempDir()
	out, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), `no gate "lint"; the gates are golangci-lint-linux, kitvet-linux,`) {
		t.Errorf("err %v\n%s", err, out)
	}
}
