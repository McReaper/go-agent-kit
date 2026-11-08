package proof_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/McReaper/rumblestrip-go/setup"
)

// WHY: pip installs zizmor outside the PATH on Windows; rumblestrip check finds it
// on the PATH like every tool, so the proof puts it there.
func withZizmor(t *testing.T) {
	t.Helper()
	if p := os.Getenv("ZIZMOR"); p != "" {
		t.Setenv("PATH", filepath.Dir(p)+string(os.PathListSeparator)+os.Getenv("PATH"))
	}
}

// WHY: an app's CI runs rumblestrip check alone; a workflow, a script or a secret
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
			t.Errorf("rumblestrip check lacks %q", want)
		}
	}
	if code != 1 {
		t.Errorf("exit %d", code)
	}
	if t.Failed() {
		t.Log(out)
	}
}

// WHY: an app keeps rumblestrip.yml and rumblestrip-approval.yml, pinned to a rumblestrip
// version; once that version needs them otherwise, a permission more or a
// file more, the app's own check says so and rumblestrip init writes them again.
func TestKitCheckFailsOnAKitWorkflowItsKitVersionWouldNotWrite(t *testing.T) {
	withZizmor(t)
	for _, c := range []struct {
		name string
		edit func(t *testing.T, dir string)
		says string
	}{
		{"a permission dropped", func(t *testing.T, dir string) {
			t.Helper()
			ci := filepath.Join(dir, ".github", "workflows", "rumblestrip.yml")
			data, err := os.ReadFile(ci)
			if err != nil {
				t.Fatal(err)
			}
			trimmed := strings.Replace(string(data), "      statuses: write # the owner's approval\n", "", 1)
			if err := os.WriteFile(ci, []byte(trimmed), 0o600); err != nil {
				t.Fatal(err)
			}
		}, ".github/workflows/rumblestrip.yml:1: rumblestrip-files: T2: it differs from the one this Rumblestrip version writes"},
		{"the approval's workflow gone", func(t *testing.T, dir string) {
			t.Helper()
			if err := os.Remove(filepath.Join(dir, ".github", "workflows", "rumblestrip-approval.yml")); err != nil {
				t.Fatal(err)
			}
		}, ".github/workflows/rumblestrip-approval.yml:1: rumblestrip-files: T2: this Rumblestrip version writes it and the repository has none"},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir := changeOf(t, trialCase{name: "no change"})
			fs, err := setup.Files(setup.Options{Version: setup.Version{Tag: "v0.2.0", Commit: strings.Repeat("ab", 20)}, Owner: "octo", OwnerEmail: "1+octo@users.noreply.github.com"})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := setup.Write(dir, fs); err != nil {
				t.Fatal(err)
			}
			if out, code := kitRun(t, dir, "check", "--only", "rumblestrip-files,actionlint,zizmor"); code != 0 {
				t.Fatalf("what rumblestrip init wrote: exit %d\n%s", code, out)
			}
			c.edit(t, dir)
			out, code := kitRun(t, dir, "check", "--only", "rumblestrip-files")
			if code != 1 || !strings.Contains(out, "FAIL rumblestrip files\n") || !strings.Contains(out, c.says) {
				t.Errorf("exit %d\n%s", code, out)
			}
		})
	}
}
