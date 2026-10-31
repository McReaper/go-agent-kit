package proof_test

import (
	"encoding/json"
	"os"
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

// WHY: the owner judges a change from its sheet, not its diff; every
// section must list exactly what the change did, or the owner approves
// what they never saw.
func TestKitReviewListsWhatTheOwnerMustRead(t *testing.T) {
	for _, c := range loadCases(t) {
		if !c.uses(reviewTool) {
			continue
		}
		t.Run(c.name, func(t *testing.T) {
			check(t, []trialCase{c}, reviewTool, reviewFindings(t, changeOf(t, c)))
		})
	}
}

func TestKitReviewWritesTheSheetForThePullRequest(t *testing.T) {
	dir := changeOf(t, caseNamed(t, "t2-review-sheet"))
	out, code := kitRun(t, dir, "review", "--base", "base")
	want, err := os.ReadFile("testdata/review-sheet.md")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(string(want), out); code != 0 || diff != "" {
		t.Errorf("exit %d, sheet (-want +got):\n%s", code, diff)
	}
	labels, _ := kitRun(t, dir, "review", "--base", "base", "--labels")
	if !strings.Contains(labels, "+tests\n") || !strings.Contains(labels, "+needs-approval\n") {
		t.Errorf("labels:\n%s", labels)
	}
	out, _ = kitRun(t, changeOf(t, trialCase{name: "no change"}), "review", "--base", "base")
	if !strings.Contains(out, "Nothing beyond the code itself") {
		t.Errorf("an empty change:\n%s", out)
	}
	labels, _ = kitRun(t, changeOf(t, trialCase{name: "no change"}), "review", "--base", "base", "--labels")
	if strings.Contains(labels, "+") {
		t.Errorf("an empty change earns no label:\n%s", labels)
	}
}
