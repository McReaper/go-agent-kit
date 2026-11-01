package mutesting_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/McReaper/go-agent-kit/internal/platform/mutesting"
)

type reportEntry struct {
	Mutator struct {
		Name string `json:"mutatorName"`
		File string `json:"originalFilePath"`
		Line int    `json:"originalStartLine"`
	} `json:"mutator"`
	Diff string `json:"diff"`
}

func escaped(t *testing.T, mutator, diff string) mutesting.Mutant {
	t.Helper()
	var e reportEntry
	e.Mutator.Name, e.Mutator.File, e.Mutator.Line, e.Diff = mutator, "a/a.go", 3, diff
	data, err := json.Marshal(map[string][]reportEntry{"escaped": {e}})
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "report.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	ms, err := mutesting.Read(dir)
	if err != nil || len(ms) != 1 {
		t.Fatalf("Read = %v, %v", ms, err)
	}
	return ms[0]
}

const hunk = "--- Original\n+++ New\n@@ -29,7 +29,7 @@\n \n"

func TestAMutantThatOnlyDropsAnErrorReturnIsMarked(t *testing.T) {
	cases := []struct {
		name    string
		mutator string
		diff    string
		drops   bool
	}{
		{
			"a wrapped error's return", "branch/if",
			hunk + " \tif err := os.MkdirAll(dir, 0o750); err != nil {\n-\t\treturn nil, fmt.Errorf(\"create: %w\", err)\n+\t\t_, _ = fmt.Errorf, err\n \t}\n", true,
		},
		{
			"a bare error's return", "branch/if",
			hunk + " \tif err != nil {\n-\t\treturn err\n+\t\t_ = err\n \t}\n", true,
		},
		{
			"a refusal under another test", "branch/if",
			hunk + " \tif errors.Is(err, fs.ErrExist) {\n-\t\treturn fmt.Errorf(\"note %s already exists\", n.ID)\n+\t\t_, _ = fmt.Errorf, n.ID\n \t}\n", false,
		},
		{
			"a block of two statements", "branch/if",
			hunk + " \tif err != nil {\n-\t\tlog(err)\n-\t\treturn err\n+\t\t_, _ = log, err\n \t}\n", false,
		},
		{
			"an expression mutant on the same line", "expression/remove",
			hunk + " \tif err != nil {\n-\t\treturn err\n+\t\treturn nil\n \t}\n", false,
		},
	}
	for _, c := range cases {
		if got := escaped(t, c.mutator, c.diff).DropsErrorReturn; got != c.drops {
			t.Errorf("%s: DropsErrorReturn = %v, want %v", c.name, got, c.drops)
		}
	}
}
