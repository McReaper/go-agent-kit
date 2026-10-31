package proof_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/McReaper/go-agent-kit/internal/platform/psscriptanalyzer"
	"github.com/McReaper/go-agent-kit/internal/platform/renovate"
)

const (
	psTool       = "psscriptanalyzer/"
	renovateTool = "renovate/"
)

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

func renovateFindings(t *testing.T, dir string) []finding {
	t.Helper()
	if _, err := os.Stat(filepath.Join(dir, "renovate.json")); err != nil {
		t.Fatalf("the trial must hold a renovate.json: %v", err)
	}
	msgs, err := renovate.Validate(t.Context(), dir, "renovate.json")
	if err != nil {
		t.Fatal(err)
	}
	findings := make([]finding, 0, len(msgs))
	for _, m := range msgs {
		findings = append(findings, finding{tool: renovateTool + "config", file: "renovate.json", text: m})
	}
	return findings
}

func TestPSScriptAnalyzerReportsEveryCase(t *testing.T) {
	proveCases(t, psTool, psFindings)
}

func TestRenovatesValidatorFindsNothingInTheTemplate(t *testing.T) {
	check(t, nil, renovateTool, renovateFindings(t, cleanDir))
}

func TestRenovatesValidatorReportsEveryCase(t *testing.T) {
	proveCases(t, renovateTool, renovateFindings)
}

func TestKitCheckFailsOnAPowerShellScriptAndARenovateConfig(t *testing.T) {
	withZizmor(t)
	dir := changeOf(t, caseNamed(t, "ps-script"))
	apply(t, caseNamed(t, "m8-renovate"), dir)
	gitIn(t, dir, "add", "-A")
	gitIn(t, dir, "commit", "-q", "-m", "a PowerShell script and a misspelled Renovate option")
	out, code := kitRun(t, dir, "check", "--base", "base", "--race=false")
	for _, want := range []string{
		"FAIL psscriptanalyzer\n", "scripts/install.ps1:7: PSAvoidUsingInvokeExpression",
		"FAIL renovate\n", "renovate.json: Invalid configuration option: minimalReleaseAge",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("kit check lacks %q", want)
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
// the kit runs must be those.
func TestTheKitRunsTheVersionsToolsTxtPins(t *testing.T) {
	data, err := os.ReadFile("../tools.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"\nPSScriptAnalyzer " + psscriptanalyzer.Version + "\n", "\nrenovate " + renovate.Version + "\n"} {
		if !strings.Contains(string(data), want) {
			t.Errorf("tools.txt lacks %q", strings.TrimSpace(want))
		}
	}
}
