package notes_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/McReaper/go-agent-kit/template/internal/notes"
	"github.com/McReaper/go-agent-kit/template/internal/platform/journal"
	"github.com/spf13/cobra"
)

func TestTaggedNotesOnTheCLI(t *testing.T) {
	dir := t.TempDir()
	m, err := notes.Open(filepath.Join(dir, "notes"), clock)
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	j, err := journal.Open(dir, clock)
	if err != nil {
		t.Fatal(err)
	}
	defer j.Close()
	var out bytes.Buffer
	cmds, err := m.Commands(j, &out)
	if err != nil {
		t.Fatal(err)
	}
	root := &cobra.Command{Use: "notes", SilenceUsage: true, SilenceErrors: true}
	root.AddCommand(cmds...)
	run := func(args ...string) error {
		root.SetArgs(args)
		out.Reset()
		return root.Execute()
	}
	if err := run("add-note", "--title", "Ship it", "--tags", "Go,go,Release"); err != nil {
		t.Fatal(err)
	}
	if err := run("list-notes", "--tag", "release"); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); !strings.Contains(got, `"tags":["go","release"]`) || strings.Contains(got, "Buy") {
		t.Errorf("list --tag release printed %s", got)
	}
}
