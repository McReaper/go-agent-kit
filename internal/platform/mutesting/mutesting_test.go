package mutesting_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/McReaper/go-agent-kit/internal/platform/mutesting"
)

func TestTheReportCountsTimeoutsAsKilled(t *testing.T) {
	dir := t.TempDir()
	data := `{
		"escaped": [{"mutator": {"mutatorName": "branch/if", "originalFilePath": "a/a.go", "originalStartLine": 3},
			"diff": "--- a.go\n+++ a.go.0\n@@ -3 +3 @@\n-\tif x {\n+\tif false {\n"}],
		"killed": [{"mutator": {"mutatorName": "expression/comparison", "originalFilePath": "a/a.go", "originalStartLine": 4},
			"diff": "-\treturn y\n"}],
		"timeouted": [{"mutator": {"mutatorName": "statement/remove", "originalFilePath": "a/a.go", "originalStartLine": 5}}]
	}`
	if err := os.WriteFile(filepath.Join(dir, "report.json"), []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := mutesting.Read(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []mutesting.Mutant{
		{File: "a/a.go", Line: 3, Mutator: "branch/if", Wrote: "if false {"},
		{File: "a/a.go", Line: 4, Mutator: "expression/comparison", Killed: true},
		{File: "a/a.go", Line: 5, Mutator: "statement/remove", Killed: true},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("mutants (-want +got):\n%s", diff)
	}
}

func TestAMissingReportIsAnError(t *testing.T) {
	if _, err := mutesting.Read(t.TempDir()); err == nil {
		t.Error("no error for a missing report")
	}
}
