package proof_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// WHY: the proof builds kit as kit.exe on every system; the hook calls
// kit, which only Windows resolves to kit.exe. The PATH gets a copy under
// the system's own name.
func kitOnPath(t *testing.T) {
	t.Helper()
	data, err := os.ReadFile(kitCommand(t, "kit"))
	if err != nil {
		t.Fatal(err)
	}
	name := "kit"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, name), data, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

// WHY: the hook is the fastest gate: it stops a secret before it enters
// the history, where only a rewrite removes it.
func TestThePreCommitHookStopsAStagedSecret(t *testing.T) {
	dir := changeOf(t, trialCase{name: "no change"})
	kitOnPath(t)
	t.Setenv("KIT_BASE", "base")
	hook := func() (string, int) {
		cmd := exec.CommandContext(t.Context(), "sh", ".githooks/pre-commit")
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			return string(out), exitCode(err)
		}
		return string(out), 0
	}
	if out, code := hook(); code != 0 {
		t.Fatalf("nothing staged: exit %d\n%s", code, out)
	}
	token := "ghp_" + strings.Repeat("a1B2c3D4e5F6", 3)
	if err := os.WriteFile(filepath.Join(dir, "settings.txt"), []byte("token = \""+token+"\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	gitIn(t, dir, "add", "settings.txt")
	if out, code := hook(); code != 1 || !strings.Contains(out, "leaks found: 1") {
		t.Errorf("a staged token: exit %d\n%s", code, out)
	}
}
