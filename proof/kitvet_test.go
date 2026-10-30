package proof_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"testing"
)

const kitvetTool = "kitvet/"

type kitvetResult struct {
	pkg      string
	analyzer string
	raw      json.RawMessage
}

type kitvetDiagnostic struct {
	Posn    string `json:"posn"`
	Message string `json:"message"`
}

func runKitvet(t *testing.T, dir, goos string) map[string]map[string]json.RawMessage {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), kitCommand(t, "kitvet"), "-json", "./...")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOOS="+goos, "GOFLAGS=-mod=readonly")
	out, err := cmd.Output()
	var exit *exec.ExitError
	if err != nil && !errors.As(err, &exit) {
		t.Fatalf("kitvet (GOOS=%s): %v", goos, err)
	}
	report := map[string]map[string]json.RawMessage{}
	if err := json.Unmarshal(out, &report); err != nil {
		t.Fatalf("kitvet (GOOS=%s): %v\n%s%s", goos, err, out, stderrOf(err))
	}
	return report
}

func kitvetFindings(t *testing.T, dir string) []finding {
	t.Helper()
	var all []finding
	for _, goos := range targetGOOS {
		for pkg, analyzers := range runKitvet(t, dir, goos) {
			for name, raw := range analyzers {
				for _, f := range decodeKitvet(t, dir, kitvetResult{pkg: pkg, analyzer: name, raw: raw}) {
					if !slices.Contains(all, f) {
						all = append(all, f)
					}
				}
			}
		}
	}
	return all
}

func decodeKitvet(t *testing.T, dir string, at kitvetResult) []finding {
	t.Helper()
	// WHY: an analyzer that couldn't run fails the proof: a skipped analysis
	// is not a clean one.
	if !bytes.HasPrefix(bytes.TrimSpace(at.raw), []byte("[")) {
		t.Fatalf("kitvet %s on %s: %s", at.analyzer, at.pkg, at.raw)
	}
	var diags []kitvetDiagnostic
	if err := json.Unmarshal(at.raw, &diags); err != nil {
		t.Fatal(err)
	}
	findings := make([]finding, 0, len(diags))
	for _, d := range diags {
		file, line := splitPosn(t, d.Posn)
		findings = append(findings, finding{tool: kitvetTool + at.analyzer, file: relTo(t, dir, file), line: line, text: d.Message})
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
		t.Fatalf("kitvet position %q is not file:line:col", posn)
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

func TestKitvetFindsNothingInTheCleanTrial(t *testing.T) {
	check(t, nil, kitvetTool, kitvetFindings(t, cleanDir))
}

func TestKitvetReportsEveryCase(t *testing.T) {
	proveCases(t, kitvetTool, kitvetFindings)
}

func TestEveryAnalyzerHasACase(t *testing.T) {
	entries, err := os.ReadDir("../analyzers")
	if err != nil {
		t.Fatal(err)
	}
	proven := map[string]bool{}
	for _, c := range loadCases(t) {
		for _, w := range c.wants {
			proven[strings.TrimPrefix(w.tool, kitvetTool)] = true
		}
	}
	for _, e := range entries {
		if e.IsDir() && e.Name() != "internal" && !proven[e.Name()] {
			t.Errorf("analyzer %s has no case under trial/cases proving it fires on the trial", e.Name())
		}
	}
}
