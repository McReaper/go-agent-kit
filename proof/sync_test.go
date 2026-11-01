package proof_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readIn(t *testing.T, dir, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(name)))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func writeIn(t *testing.T, dir, name, text string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, filepath.FromSlash(name)), []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
}

// WHY: after gonew, an app's copies of the kit's skills, hooks, workflow
// pin and AGENTS.md rules drift with no way back; kit sync brings them to a
// kit version, leaves the app's own words alone, and its diff is reviewed.
func TestKitSyncBringsAnAppsKitFilesToAVersionAndKeepsItsOwnWords(t *testing.T) {
	app := overlay(t)
	tmpl, err := filepath.Abs(cleanDir)
	if err != nil {
		t.Fatal(err)
	}
	skill := ".claude/skills/new-feature/SKILL.md"
	writeIn(t, app, skill, "an agent's edit of the skill\n")
	agents := readIn(t, app, "AGENTS.md")
	appWords := strings.Replace(agents, "notes keeps short notes", "notes, renamed by the app, keeps short notes", 1)
	writeIn(t, app, "AGENTS.md", strings.Replace(appWords, "- Code only tests reach: delete it with its tests.\n", "", 1))
	version := strings.Repeat("ab", 20)
	out, code := kitRun(t, app, "sync", "--from", tmpl, "--version", version)
	if code != 0 || !strings.Contains(out, "updated "+skill) || !strings.Contains(out, "updated AGENTS.md") {
		t.Fatalf("kit sync: exit %d\n%s", code, out)
	}
	if readIn(t, app, skill) != readIn(t, tmpl, skill) {
		t.Error("the skill is not the kit's")
	}
	if got := readIn(t, app, "AGENTS.md"); got != appWords {
		t.Errorf("AGENTS.md: the app's words or the kit's section are wrong:\n%s", got)
	}
	if kitYML := readIn(t, app, ".github/workflows/kit.yml"); strings.Count(kitYML, version) != 2 {
		t.Errorf("kit.yml is not pinned twice to %s:\n%s", version, kitYML)
	}
}
