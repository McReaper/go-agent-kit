package proof_test

import (
	"encoding/json"
	"testing"
)

const sweepTool = "sweep/"

func sweepFindings(t *testing.T, dir string) []finding {
	t.Helper()
	out := runIn(t, dir, []int{1}, kitCommand(t, "kit"), "sweep", "--json")
	var rows []guardRow
	if err := json.Unmarshal(out, &rows); err != nil {
		t.Fatalf("kit sweep: %v\n%s", err, out)
	}
	findings := make([]finding, 0, len(rows))
	for _, r := range rows {
		findings = append(findings, finding{tool: sweepTool + r.Rule, file: r.File, line: r.Line, text: r.Message})
	}
	return findings
}

func TestSweepFindsNothingInTheCleanTrial(t *testing.T) {
	check(t, nil, sweepTool, sweepFindings(t, cleanDir))
}

func TestSweepReportsEveryCase(t *testing.T) {
	proveCases(t, sweepTool, sweepFindings)
}
