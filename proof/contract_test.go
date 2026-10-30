package proof_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// WHY: a library's exported API is a contract (E5): apidiff compares the
// clean trial's, the base, with the change's, and reports what breaks callers.
func apidiffFindings(t *testing.T, dir string) []finding {
	t.Helper()
	var findings []finding
	for pkg := range strings.FieldsSeq(string(runIn(t, cleanDir, nil, "go", "list", "./pkg/..."))) {
		base := filepath.Join(t.TempDir(), "base.export")
		runIn(t, cleanDir, nil, "apidiff", "-w", base, pkg)
		out := runIn(t, dir, nil, "apidiff", "-incompatible", base, pkg)
		for line := range strings.Lines(string(out)) {
			if change, ok := strings.CutPrefix(strings.TrimSpace(line), "- "); ok {
				file := strings.TrimPrefix(pkg, "example.com/notes/")
				findings = append(findings, finding{tool: "apidiff/incompatible", file: file, line: 0, text: change})
			}
		}
	}
	return findings
}

func TestApidiffFindsNothingInTheCleanTrial(t *testing.T) {
	check(t, nil, "apidiff/", apidiffFindings(t, cleanDir))
}

func TestApidiffReportsEveryCase(t *testing.T) {
	proveCases(t, "apidiff/", apidiffFindings)
}

type sbom struct {
	BOMFormat  string `json:"bomFormat"`
	Components []struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"components"`
}

// WHY: cyclonedx-gomod runs the go on PATH; the module's own toolchain goes
// first, or an older go fails to read the standard library's module.
func TestSBOMListsTheModulesTheBinaryLinks(t *testing.T) {
	goroot := strings.TrimSpace(string(runIn(t, cleanDir, nil, "go", "env", "GOROOT")))
	out := filepath.Join(t.TempDir(), "bom.json")
	cmd := exec.CommandContext(t.Context(), "cyclonedx-gomod", "app", "-json", "-main", "cmd/notes", "-output", out, ".")
	cmd.Dir = cleanDir
	cmd.Env = append(os.Environ(), "PATH="+filepath.Join(goroot, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"))
	if msg, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("cyclonedx-gomod: %v\n%s", err, msg)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var b sbom
	if err := json.Unmarshal(data, &b); err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, c := range b.Components {
		names = append(names, c.Name+"@"+c.Version)
	}
	for _, want := range []string{"github.com/modelcontextprotocol/go-sdk@v1.8.0", "golang.org/x/text@v0.41.0"} {
		if !slices.Contains(names, want) {
			t.Errorf("SBOM lacks %s: %v", want, names)
		}
	}
	if slices.ContainsFunc(names, func(n string) bool { return strings.HasPrefix(n, "pgregory.net/rapid") }) {
		t.Errorf("SBOM lists rapid, which only tests use: %v", names)
	}
}
