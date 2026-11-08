package proof_test

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const reviewTool = "review/"

type reviewRow struct {
	Section string `json:"section"`
	File    string `json:"file"`
	Line    int    `json:"line"`
	Text    string `json:"text"`
}

func reviewFindings(t *testing.T, dir string) []finding {
	t.Helper()
	out, code := kitRun(t, dir, "review", "--base", "base", "--json")
	if code != 0 {
		t.Fatalf("kit review: exit %d\n%s", code, out)
	}
	var rows []reviewRow
	if err := json.Unmarshal([]byte(out), &rows); err != nil {
		t.Fatalf("kit review: %v\n%s", err, out)
	}
	findings := make([]finding, 0, len(rows))
	for _, r := range rows {
		findings = append(findings, finding{tool: reviewTool + r.Section, file: r.File, line: r.Line, text: r.Text})
	}
	return findings
}

// WHY: the owner judges a change from its sheet, not its diff: it must
// list exactly what is theirs, the spec first, or they approve what they
// never saw; and nothing else, or they stop reading it.
func TestKitReviewListsWhatTheOwnerMustRead(t *testing.T) {
	for _, c := range loadCases(t) {
		if !c.uses(reviewTool) {
			continue
		}
		t.Run(c.name, func(t *testing.T) {
			check(t, wantsOf(reviewTool, c), reviewFindings(t, changeOf(t, c)))
		})
	}
}

// WHY: the proofs' build of kit stamps the sheet's footer with a version
// from the checkout, which moves with every commit.
var footerVersion = regexp.MustCompile(`\) v[^ <]+</sub>`)

func TestKitReviewWritesTheSheetForThePullRequest(t *testing.T) {
	dir := changeOf(t, caseNamed(t, "t2-review-sheet"))
	out, code := kitRun(t, dir, "review", "--base", "base")
	want, err := os.ReadFile("testdata/review-sheet.md")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(string(want), footerVersion.ReplaceAllString(out, ") VERSION</sub>")); code != 0 || diff != "" {
		t.Errorf("exit %d, sheet (-want +got):\n%s", code, diff)
	}
	labels, _ := kitRun(t, dir, "review", "--base", "base", "--labels")
	if labels != "+human-review\n" {
		t.Errorf("labels:\n%s", labels)
	}
	out, _ = kitRun(t, changeOf(t, trialCase{name: "no change"}), "review", "--base", "base")
	if !strings.Contains(out, "Nothing for you: the kit merges it once they pass.") {
		t.Errorf("an empty change:\n%s", out)
	}
	labels, _ = kitRun(t, changeOf(t, trialCase{name: "no change"}), "review", "--base", "base", "--labels")
	if strings.Contains(labels, "+") {
		t.Errorf("an empty change earns no label:\n%s", labels)
	}
}

// WHY: the review sheet's job posts the owner's approval status the
// ruleset requires, and the label, from where their approval stands: once
// they approved the commit, both say so.
func TestKitReviewSaysWhereTheOwnersApprovalStands(t *testing.T) {
	dir := changeOf(t, caseNamed(t, "t2-review-sheet"))
	for _, c := range []struct {
		args            []string
		approval, label string
	}{
		{nil, "waiting\n", "+human-review\n"},
		{[]string{"--approved"}, "approved\n", "-human-review\n"},
	} {
		approval, _ := kitRun(t, dir, append([]string{"review", "--base", "base", "--approval"}, c.args...)...)
		labels, _ := kitRun(t, dir, append([]string{"review", "--base", "base", "--labels"}, c.args...)...)
		if approval != c.approval || labels != c.label {
			t.Errorf("%v: approval %q, labels %q; want %q, %q", c.args, approval, labels, c.approval, c.label)
		}
	}
	if approval, _ := kitRun(t, changeOf(t, trialCase{name: "no change"}), "review", "--base", "base", "--approval"); approval != "none\n" {
		t.Errorf("an empty change: approval %q", approval)
	}
}
