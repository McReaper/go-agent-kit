package proof_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/McReaper/rumblestrip-go/analyzers/all"
)

const vetTool = "rumblestrip-vet/"

type vetResult struct {
	pkg      string
	analyzer string
	raw      json.RawMessage
}

type vetDiagnostic struct {
	Posn    string `json:"posn"`
	Message string `json:"message"`
}

func runVet(t *testing.T, dir, goos string) map[string]map[string]json.RawMessage {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), kitCommand(t, "rumblestrip-vet"), "-json", "./...")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOOS="+goos, "GOFLAGS=-mod=readonly")
	out, err := cmd.Output()
	if err != nil && exitCode(err) < 0 {
		t.Fatalf("rumblestrip-vet (GOOS=%s): %v", goos, err)
	}
	report := map[string]map[string]json.RawMessage{}
	if err := json.Unmarshal(out, &report); err != nil {
		t.Fatalf("rumblestrip-vet (GOOS=%s): %v\n%s%s", goos, err, out, stderrOf(err))
	}
	return report
}

func vetFindings(t *testing.T, dir string) []finding {
	t.Helper()
	var all []finding
	for _, goos := range targetGOOS {
		for pkg, analyzers := range runVet(t, dir, goos) {
			for name, raw := range analyzers {
				for _, f := range decodeVet(t, dir, vetResult{pkg: pkg, analyzer: name, raw: raw}) {
					if !slices.Contains(all, f) {
						all = append(all, f)
					}
				}
			}
		}
	}
	return all
}

func decodeVet(t *testing.T, dir string, at vetResult) []finding {
	t.Helper()
	// WHY: an analyzer that couldn't run fails the proof: a skipped analysis
	// is not a clean one.
	if !bytes.HasPrefix(bytes.TrimSpace(at.raw), []byte("[")) {
		t.Fatalf("rumblestrip-vet %s on %s: %s", at.analyzer, at.pkg, at.raw)
	}
	var diags []vetDiagnostic
	if err := json.Unmarshal(at.raw, &diags); err != nil {
		t.Fatal(err)
	}
	findings := make([]finding, 0, len(diags))
	for _, d := range diags {
		file, line := splitPosn(t, d.Posn)
		findings = append(findings, finding{tool: vetTool + at.analyzer, file: relTo(t, dir, file), line: line, text: d.Message})
	}
	return findings
}

// WHY: split from the right: a Windows file name holds a colon after its
// drive letter.
func splitPosn(t *testing.T, posn string) (string, int) {
	t.Helper()
	rest, _, ok := cutLast(posn, ":")
	file, lineStr, ok2 := cutLast(rest, ":")
	line, err := strconv.Atoi(lineStr)
	if !ok || !ok2 || err != nil {
		t.Fatalf("rumblestrip-vet position %q is not file:line:col", posn)
	}
	return file, line
}

func cutLast(s, sep string) (string, string, bool) {
	i := strings.LastIndex(s, sep)
	if i < 0 {
		return s, "", false
	}
	return s[:i], s[i+len(sep):], true
}

func TestVetFindsNothingInTheCleanTrial(t *testing.T) {
	check(t, wantsOf(vetTool), vetFindings(t, cleanDir))
}

func TestVetReportsEveryCase(t *testing.T) {
	proveCases(t, vetTool, vetFindings)
}

func TestEveryAnalyzerHasACase(t *testing.T) {
	proven := map[string]bool{}
	for _, c := range loadCases(t) {
		for _, w := range c.wants {
			proven[strings.TrimPrefix(w.tool, vetTool)] = true
		}
	}
	for _, a := range all.Analyzers() {
		if !proven[a.Name] {
			t.Errorf("analyzer %s has no case under trial/cases proving it fires on the trial", a.Name)
		}
	}
}
