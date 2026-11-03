package proof_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/McReaper/go-agent-kit/kitinit"
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

// WHY: an app keeps kit.yml, pinned to a kit version; once that version
// needs another kit.yml, a trigger or a permission more, the app's own
// check says so and kit init writes it again.
func TestKitCheckFailsOnAKitYMLItsKitVersionWouldNotWrite(t *testing.T) {
	withZizmor(t)
	dir := changeOf(t, trialCase{name: "no change"})
	fs, err := kitinit.Files(kitinit.Options{Version: kitinit.Version{Tag: "v0.2.0", Commit: strings.Repeat("ab", 20)}, Owner: "octo"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := kitinit.Write(dir, fs); err != nil {
		t.Fatal(err)
	}
	ci := filepath.Join(dir, ".github", "workflows", "kit.yml")
	data, err := os.ReadFile(ci)
	if err != nil {
		t.Fatal(err)
	}
	trimmed := strings.Replace(string(data), "  pull_request_review:\n    types: [submitted]\n", "", 1)
	if err := os.WriteFile(ci, []byte(trimmed), 0o600); err != nil {
		t.Fatal(err)
	}
	gitIn(t, dir, "add", "-A")
	gitIn(t, dir, "commit", "-q", "-m", "kit init, then a trigger dropped")
	out, code := kitRun(t, dir, "check", "--base", "base", "--race=false")
	if code != 1 || !strings.Contains(out, "FAIL kit files\n") || !strings.Contains(out, "it differs from the kit.yml this kit version writes") {
		t.Errorf("exit %d\n%s", code, out)
	}
}
