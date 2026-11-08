package proof_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func bash(t *testing.T, dir, command string) string {
	t.Helper()
	data, err := json.Marshal(map[string]any{"tool_name": "Bash", "cwd": dir, "tool_input": map[string]string{"command": command}})
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// WHY: the plugin asks rumblestrip before each command the agent runs. A gesture
// past Rumblestrip is refused before it reaches the platform, and a commit
// waits until it holds no secret: one committed stays in the history,
// where only a rewrite removes it.
func TestThePreToolHookRefusesAGestureAndACommitWithAStagedSecret(t *testing.T) {
	dir := changeOf(t, trialCase{name: "no change"})
	if out, code := hookRun(t, bash(t, dir, "git -C . push origin main"), "pre-tool"); code != 2 || !strings.Contains(out, "through a pull request only") {
		t.Errorf("a push to main: exit %d\n%s", code, out)
	}
	if out, code := hookRun(t, bash(t, dir, "git commit -m notes"), "pre-tool"); code != 0 {
		t.Fatalf("a commit with nothing wrong staged: exit %d\n%s", code, out)
	}
	token := "ghp_" + strings.Repeat("a1B2c3D4e5F6", 3)
	if err := os.WriteFile(filepath.Join(dir, "settings.txt"), []byte("token = \""+token+"\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitIn(t, dir, "add", "settings.txt")
	if out, code := hookRun(t, bash(t, dir, "git add -A && git commit -m notes"), "pre-tool"); code != 2 || !strings.Contains(out, "settings.txt:1: github-pat") {
		t.Errorf("a commit with a staged token: exit %d\n%s", code, out)
	}
}
