package proof_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const specTool = "spec/"

func specFindings(t *testing.T, dir string) []finding {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), kitCommand(t, "rumblestrip"), "spec", "--json")
	cmd.Dir = dir
	cmd.Env = os.Environ()
	out, err := cmd.Output()
	if err != nil && exitCode(err) != 1 {
		t.Fatalf("rumblestrip spec: %v\n%s", err, stderrOf(err))
	}
	var rows []guardRow
	if err := json.Unmarshal(out, &rows); err != nil {
		t.Fatalf("rumblestrip spec: %v\n%s", err, out)
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

// WHY: steps that don't compile are the change's to fix: the spec reports
// the compiler's error at its line, as a finding the agent fixes, never as
// a gate that could not run, which reads as the kit's fault.
func TestKitSpecReportsStepsThatDontCompileAsTheChanges(t *testing.T) {
	dir := changeOf(t, trialCase{name: "no change"})
	steps := filepath.Join(dir, "features", "notes_test.go")
	data, err := os.ReadFile(steps)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(steps, append(data, []byte("\nfunc broken() string { return \"unterminated }\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
	var built []finding
	for _, f := range specFindings(t, dir) {
		if f.tool == specTool+"spec-build" {
			built = append(built, f)
		}
	}
	if len(built) == 0 || built[0].file != "features/notes_test.go" || built[0].line == 0 || !strings.Contains(built[0].text, "the spec's steps don't compile") {
		t.Errorf("findings %+v, want spec-build at features/notes_test.go", built)
	}
}
