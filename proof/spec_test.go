package proof_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"testing"
)

const specTool = "spec/"

func specFindings(t *testing.T, dir string) []finding {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), kitCommand(t, "kit"), "spec", "--json")
	cmd.Dir = dir
	cmd.Env = os.Environ()
	out, err := cmd.Output()
	if err != nil && exitCode(err) != 1 {
		t.Fatalf("kit spec: %v\n%s", err, stderrOf(err))
	}
	var rows []guardRow
	if err := json.Unmarshal(out, &rows); err != nil {
		t.Fatalf("kit spec: %v\n%s", err, out)
	}
	findings := make([]finding, 0, len(rows))
	for _, r := range rows {
		findings = append(findings, finding{tool: specTool + r.Rule, file: r.File, line: r.Line, text: r.Message})
	}
	return findings
}

// WHY: the owner approves scenarios, not the steps that run them; a step
// that passes whatever it is told to expect would make the approval a
// sentence no program honours.
func TestKitSpecReportsEveryCase(t *testing.T) {
	for _, c := range loadCases(t) {
		if !c.uses(specTool) {
			continue
		}
		t.Run(c.name, func(t *testing.T) {
			check(t, wantsOf(specTool, c), specFindings(t, changeOf(t, c)))
		})
	}
	t.Run("template", func(t *testing.T) {
		check(t, wantsOf(specTool), specFindings(t, changeOf(t, trialCase{name: "no change"})))
	})
}
