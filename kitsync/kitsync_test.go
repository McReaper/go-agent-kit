package kitsync_test

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/McReaper/go-agent-kit/kitsync"
	"github.com/google/go-cmp/cmp"
)

const version = "0123456789abcdef0123456789abcdef01234567"

func write(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, text := range files {
		p := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func read(t *testing.T, dir, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(name)))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func template(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	write(t, dir, map[string]string{
		kitsync.Manifest:            "# owned\n.claude/skills/feature/\n.github/workflows/kit.yml\n",
		".claude/skills/feature/a":  "skill a, as the kit ships it\n",
		".claude/skills/feature/b":  "skill b\n",
		".github/workflows/kit.yml": "uses: McReaper/go-agent-kit/.github/workflows/kit-check.yml@" + strings.Repeat("9", 40) + "\n  kit-version: " + strings.Repeat("9", 40) + "\n",
		kitsync.Agents:              "# the template's app\n\n<!-- kit:begin -->\nthe kit's rules, new\n<!-- kit:end -->\n",
	})
	return dir
}

func TestSyncBringsTheKitsFilesAndKeepsTheAppsOwn(t *testing.T) {
	app := t.TempDir()
	write(t, app, map[string]string{
		".claude/skills/feature/a":     "skill a, edited in the app\n",
		".claude/skills/feature/stale": "a file the kit dropped\n",
		".github/workflows/kit.yml":    "an old pin\n",
		kitsync.Agents:                 "# notes, the app's own words\n\n<!-- kit:begin -->\nthe kit's rules, old\n<!-- kit:end -->\n\n## The app's own section\n",
		"main.go":                      "package main\n",
	})
	changes, err := kitsync.Sync(app, template(t), version)
	if err != nil {
		t.Fatal(err)
	}
	want := []kitsync.Change{
		{".claude/skills/feature/a", "updated"},
		{".claude/skills/feature/b", "added"},
		{".claude/skills/feature/stale", "removed"},
		{".github/workflows/kit.yml", "updated"},
		{kitsync.Manifest, "added"},
		{kitsync.Agents, "updated"},
	}
	if diff := cmp.Diff(want, changes); diff != "" {
		t.Errorf("changes (-want +got):\n%s", diff)
	}
	if got := read(t, app, ".claude/skills/feature/a"); got != "skill a, as the kit ships it\n" {
		t.Errorf("skill a = %q", got)
	}
	if _, err := os.Stat(filepath.Join(app, ".claude/skills/feature/stale")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the dropped file is still there: %v", err)
	}
	wantAgents := "# notes, the app's own words\n\n<!-- kit:begin -->\nthe kit's rules, new\n<!-- kit:end -->\n\n## The app's own section\n"
	if diff := cmp.Diff(wantAgents, read(t, app, kitsync.Agents)); diff != "" {
		t.Errorf("AGENTS.md (-want +got):\n%s", diff)
	}
	if got := read(t, app, "main.go"); got != "package main\n" {
		t.Errorf("a file the kit doesn't own changed: %q", got)
	}
}

func TestSyncPinsTheWorkflowToTheVersionAndSyncsNothingTwice(t *testing.T) {
	app, tmpl := t.TempDir(), template(t)
	if _, err := kitsync.Sync(app, tmpl, version); err != nil {
		t.Fatal(err)
	}
	want := "uses: McReaper/go-agent-kit/.github/workflows/kit-check.yml@" + version + "\n  kit-version: " + version + "\n"
	if got := read(t, app, ".github/workflows/kit.yml"); got != want {
		t.Errorf("kit.yml = %q, want both pins on %s", got, version)
	}
	again, err := kitsync.Sync(app, tmpl, version)
	if err != nil || len(again) != 0 {
		t.Errorf("a second sync changed %v (%v), want nothing", again, err)
	}
}

func TestAnAppWithoutTheKitsMarkersGetsTheSectionAtItsEnd(t *testing.T) {
	app := t.TempDir()
	write(t, app, map[string]string{kitsync.Agents: "# notes\n\nThe app's own words.\n"})
	if _, err := kitsync.Sync(app, template(t), version); err != nil {
		t.Fatal(err)
	}
	want := "# notes\n\nThe app's own words.\n\n<!-- kit:begin -->\nthe kit's rules, new\n<!-- kit:end -->\n"
	if diff := cmp.Diff(want, read(t, app, kitsync.Agents)); diff != "" {
		t.Errorf("AGENTS.md (-want +got):\n%s", diff)
	}
}

func TestSyncRefusesAVersionThatIsntAFullCommitID(t *testing.T) {
	for _, v := range []string{"", "main", "v1.2.3", version[:12], strings.ToUpper(version)} {
		if _, err := kitsync.Sync(t.TempDir(), template(t), v); !errors.Is(err, kitsync.ErrVersion) {
			t.Errorf("Sync(%q): err = %v, want ErrVersion", v, err)
		}
	}
}

// WHY: the module cache holds the template read-only, and a first kit sync
// copied that mode: the app could no longer edit the files it got, nor a
// second sync rewrite them.
func TestSyncedFilesStayWritableAndScriptsRunWhateverTheTemplatesMode(t *testing.T) {
	tmpl, app := template(t), t.TempDir()
	write(t, tmpl, map[string]string{".claude/skills/feature/hook": "#!/bin/sh\necho hook\n"})
	readOnly := func(dir string) {
		for _, name := range []string{".claude/skills/feature/a", ".claude/skills/feature/hook", kitsync.Agents} {
			if err := os.Chmod(filepath.Join(dir, filepath.FromSlash(name)), 0o400); err != nil {
				t.Fatal(err)
			}
		}
	}
	readOnly(tmpl)
	if _, err := kitsync.Sync(app, tmpl, version); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{".claude/skills/feature/a", ".claude/skills/feature/hook", kitsync.Agents} {
		info, err := os.Stat(filepath.Join(app, filepath.FromSlash(name)))
		if err != nil || info.Mode().Perm()&0o200 == 0 {
			t.Errorf("%s: mode %v (%v), want writable", name, info.Mode(), err)
		}
	}
	if info, err := os.Stat(filepath.Join(app, ".claude/skills/feature/hook")); runtime.GOOS != "windows" && (err != nil || info.Mode().Perm()&0o100 == 0) {
		t.Errorf("the script: mode %v (%v), want executable", info.Mode(), err)
	}
	readOnly(app)
	if err := os.Chmod(filepath.Join(tmpl, ".claude/skills/feature/a"), 0o600); err != nil {
		t.Fatal(err)
	}
	write(t, tmpl, map[string]string{".claude/skills/feature/a": "skill a, the kit's next version\n"})
	if _, err := kitsync.Sync(app, tmpl, version); err != nil {
		t.Fatalf("a second sync over the read-only files: %v", err)
	}
	if got := read(t, app, ".claude/skills/feature/a"); got != "skill a, the kit's next version\n" {
		t.Errorf("skill a = %q", got)
	}
}
