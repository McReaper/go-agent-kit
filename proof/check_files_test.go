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

// WHY: an app keeps kit.yml and kit-approval.yml, pinned to a kit
// version; once that version needs them otherwise, a permission more or a
// file more, the app's own check says so and kit init writes them again.
func TestKitCheckFailsOnAKitWorkflowItsKitVersionWouldNotWrite(t *testing.T) {
	withZizmor(t)
	for _, c := range []struct {
		name string
		edit func(t *testing.T, dir string)
		says string
	}{
		{"a permission dropped", func(t *testing.T, dir string) {
			t.Helper()
			ci := filepath.Join(dir, ".github", "workflows", "kit.yml")
			data, err := os.ReadFile(ci)
			if err != nil {
				t.Fatal(err)
			}
			trimmed := strings.Replace(string(data), "      statuses: write # the owner's approval\n", "", 1)
			if err := os.WriteFile(ci, []byte(trimmed), 0o600); err != nil {
				t.Fatal(err)
			}
		}, ".github/workflows/kit.yml:1: kit-files: T2: it differs from the one this kit version writes"},
		{"the approval's workflow gone", func(t *testing.T, dir string) {
			t.Helper()
			if err := os.Remove(filepath.Join(dir, ".github", "workflows", "kit-approval.yml")); err != nil {
				t.Fatal(err)
			}
		}, ".github/workflows/kit-approval.yml:1: kit-files: T2: this kit version writes it and the repository has none"},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir := changeOf(t, trialCase{name: "no change"})
			fs, err := kitinit.Files(kitinit.Options{Version: kitinit.Version{Tag: "v0.2.0", Commit: strings.Repeat("ab", 20)}, Owner: "octo", OwnerEmail: "1+octo@users.noreply.github.com"})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := kitinit.Write(dir, fs); err != nil {
				t.Fatal(err)
			}
			if out, code := kitRun(t, dir, "check", "--only", "kit-files,actionlint,zizmor"); code != 0 {
				t.Fatalf("what kit init wrote: exit %d\n%s", code, out)
			}
			c.edit(t, dir)
			out, code := kitRun(t, dir, "check", "--only", "kit-files")
			if code != 1 || !strings.Contains(out, "FAIL kit files\n") || !strings.Contains(out, c.says) {
				t.Errorf("exit %d\n%s", code, out)
			}
		})
	}
}
