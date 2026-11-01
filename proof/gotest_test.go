package proof_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"strings"
	"testing"
)

const (
	testTool       = "test/"
	templateModule = "github.com/McReaper/go-agent-kit/template"
)

var testFailure = regexp.MustCompile(`^\s+(\w+_test\.go):(\d+): (.*)`)

type testEvent struct {
	Action  string `json:"Action"`
	Package string `json:"Package"`
	Test    string `json:"Test"`
	Output  string `json:"Output"`
}

// WHY: some rules live in the template's own tests (the smoke test, the
// journeys), not in a linter; a case breaks the program, and each failing
// test is a finding, at the line of the check that failed. A passing test's
// logs are not findings.
func testFindings(t *testing.T, dir string) []finding {
	t.Helper()
	out := runIn(t, dir, []int{1}, "go", "test", "-json", "-count=1", "./...")
	logged := map[string][]finding{}
	var findings []finding
	dec := json.NewDecoder(bytes.NewReader(out))
	for {
		var e testEvent
		err := dec.Decode(&e)
		if errors.Is(err, io.EOF) {
			return findings
		}
		if err != nil {
			t.Fatalf("go test -json: %v", err)
		}
		key := e.Package + " " + e.Test
		if e.Action == "fail" {
			findings = append(findings, logged[key]...)
		}
		if m := testFailure.FindStringSubmatch(e.Output); m != nil && e.Test != "" {
			pkg := strings.TrimPrefix(strings.TrimPrefix(e.Package, templateModule), "/")
			logged[key] = append(logged[key], finding{tool: testTool + e.Test, file: strings.TrimPrefix(pkg+"/"+m[1], "/"), line: atoi(t, m[2]), text: m[3]})
		}
	}
}

func TestTheTemplatesTestsReportEveryCase(t *testing.T) {
	proveCases(t, testTool, testFindings)
}
