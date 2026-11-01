package proof_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// WHY: pip installs zizmor outside the PATH on Windows; kit check finds it
// on the PATH like every tool, so the proof puts it there.
func withZizmor(t *testing.T) {
	t.Helper()
	if p := os.Getenv("ZIZMOR"); p != "" {
		t.Setenv("PATH", filepath.Dir(p)+string(os.PathListSeparator)+os.Getenv("PATH"))
	}
}

// WHY: an app's CI runs kit check alone; a workflow, a script or a secret
// the Go gates never read must fail it too.
func TestKitCheckFailsOnAWorkflowAScriptAndASecret(t *testing.T) {
	withZizmor(t)
	dir := changeOf(t, caseNamed(t, "ci-workflow"))
	for _, name := range []string{"sh-script", "s1-secrets"} {
		apply(t, caseNamed(t, name), dir)
	}
	gitIn(t, dir, "add", "-A")
	gitIn(t, dir, "commit", "-q", "-m", "a workflow, a script and a secret")
	out, code := kitRun(t, dir, "check", "--base", "base", "--race=false")
	for _, want := range []string{
		"FAIL actionlint\n", "is potentially untrusted",
		"FAIL zizmor\n", "template-injection: code injection via template expansion",
		"FAIL shellcheck\n", "SC2086",
		"FAIL gitleaks\n", "internal/notes/internal/add/config.go:3: generic-api-key",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("kit check lacks %q", want)
		}
	}
	if code != 1 {
		t.Errorf("exit %d", code)
	}
	if t.Failed() {
		t.Log(out)
	}
}
