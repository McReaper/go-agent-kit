package proof_test

import (
	"testing"

	"github.com/McReaper/go-agent-kit/archtest"
)

const archTool = "archtest/"

func archFindings(t *testing.T, dir string) []finding {
	t.Helper()
	vs, err := archtest.Check(dir)
	if err != nil {
		t.Fatal(err)
	}
	findings := make([]finding, 0, len(vs))
	for _, v := range vs {
		findings = append(findings, finding{tool: archTool + v.Rule, file: relTo(t, dir, v.File), line: v.Line, text: v.Message})
	}
	return findings
}

func TestArchtestFindsNothingInTheCleanTrial(t *testing.T) {
	check(t, nil, archTool, archFindings(t, cleanDir))
}

func TestArchtestReportsEveryCase(t *testing.T) {
	proveCases(t, archTool, archFindings)
}
