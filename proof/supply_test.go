package proof_test

import (
	"os"
	"strings"
	"testing"

	"github.com/McReaper/rumblestrip-go/internal/platform/psscriptanalyzer"
)

const psTool = "psscriptanalyzer/"

func psFindings(t *testing.T, dir string) []finding {
	t.Helper()
	scripts := filesWithExt(t, dir, ".ps1", ".psm1", ".psd1")
	if len(scripts) == 0 {
		return nil
	}
	fs, err := psscriptanalyzer.Analyze(t.Context(), dir, scripts)
	if err != nil {
		t.Fatal(err)
	}
	findings := make([]finding, 0, len(fs))
	for _, f := range fs {
		findings = append(findings, finding{tool: psTool + f.Rule, file: f.File, line: f.Line, text: f.Message})
	}
	return findings
}

func TestPSScriptAnalyzerReportsEveryCase(t *testing.T) {
	proveCases(t, psTool, psFindings)
}

func TestKitCheckFailsOnAPowerShellScript(t *testing.T) {
	withZizmor(t)
	dir := changeOf(t, caseNamed(t, "ps-script"))
	out, code := kitRun(t, dir, "check", "--base", "base", "--race=false")
	for _, want := range []string{
		"FAIL psscriptanalyzer\n", "scripts/install.ps1:7: PSAvoidUsingInvokeExpression",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("rumblestrip check lacks %q", want)
		}
	}
	if code != 1 {
		t.Errorf("exit %d", code)
	}
	if t.Failed() {
		t.Log(out)
	}
}

// WHY: tools.txt is what CI installs and what the doc names; the versions
// Rumblestrip runs must be those.
func TestTheKitRunsTheVersionsToolsTxtPins(t *testing.T) {
	data, err := os.ReadFile("../tools.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"\nPSScriptAnalyzer " + psscriptanalyzer.Version + "\n"} {
		if !strings.Contains(string(data), want) {
			t.Errorf("tools.txt lacks %q", strings.TrimSpace(want))
		}
	}
}
