package notes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/McReaper/go-agent-kit/template/internal/notes"
	"github.com/McReaper/go-agent-kit/template/internal/platform/journal"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func commands(t *testing.T) map[string]*cobra.Command {
	t.Helper()
	dir := t.TempDir()
	m, err := notes.Open(filepath.Join(dir, "notes"), clock)
	if err != nil {
		t.Fatal(err)
	}
	j, err := journal.Open(dir, clock)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := errors.Join(m.Close(), j.Close()); err != nil {
			t.Error(err)
		}
	})
	cmds, err := m.Commands(j, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]*cobra.Command{}
	for _, c := range cmds {
		byName[c.Name()] = c
	}
	return byName
}

// WHY: one Args struct gives both surfaces their arguments. A field the CLI
// can't bind breaks every command at startup while the MCP tests pass (a
// crash test shipped one); a tool without its command, or a flag named
// otherwise, splits the two surfaces.
func TestEveryToolIsACommandWithTheSameArguments(t *testing.T) {
	res, err := session(t, t.TempDir()).ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	cmds := commands(t)
	for _, tool := range res.Tools {
		name := strings.ReplaceAll(tool.Name, "_", "-")
		cmd, ok := cmds[name]
		if !ok {
			t.Errorf("tool %s has no command %s", tool.Name, name)
			continue
		}
		delete(cmds, name)
		var flags []string
		cmd.Flags().VisitAll(func(f *pflag.Flag) { flags = append(flags, f.Name) })
		if diff := cmp.Diff(arguments(t, tool.InputSchema), flags); diff != "" {
			t.Errorf("%s: flags (-tool arguments +command flags):\n%s", tool.Name, diff)
		}
	}
	for name := range cmds {
		t.Errorf("command %s has no tool", name)
	}
}

func arguments(t *testing.T, schema any) []string {
	t.Helper()
	data, err := json.Marshal(schema)
	if err != nil {
		t.Fatal(err)
	}
	var s struct {
		Properties map[string]any `json:"properties"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, p := range slices.Sorted(maps.Keys(s.Properties)) {
		out = append(out, strings.ReplaceAll(p, "_", "-"))
	}
	return out
}
