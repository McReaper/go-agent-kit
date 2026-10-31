package proof_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
)

var goFence = regexp.MustCompile("(?s)```go ([^\n]+)\n(.*?)```")

// WHY: an agent copies a skill's code as written; it is proven like the
// template's own, or the agent copies code no gate has seen.
func TestTheSecretsSkillsCodePassesEveryGate(t *testing.T) {
	doc, err := os.ReadFile(filepath.Join(cleanDir, ".claude", "skills", "secrets", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	blocks := goFence.FindAllStringSubmatch(string(doc), -1)
	if len(blocks) != 2 {
		t.Fatalf("the skill has %d Go files, want the package and its test", len(blocks))
	}
	dir := overlay(t)
	var files []string
	for _, b := range blocks {
		file := filepath.Join(dir, filepath.FromSlash(b[1]))
		if err := os.MkdirAll(filepath.Dir(file), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(b[2]), 0o600); err != nil {
			t.Fatal(err)
		}
		files = append(files, file)
	}
	if out, err := exec.CommandContext(t.Context(), "gofmt", append([]string{"-l"}, files...)...).Output(); err != nil || len(out) > 0 {
		t.Errorf("gofmt: %v: %s", err, out)
	}
	runIn(t, dir, nil, "go", "test", "-count=1", "./internal/platform/secret/")
	check(t, nil, golangciTool, golangciFindings(t, dir))
	check(t, nil, kitvetTool, kitvetFindings(t, dir))
}
