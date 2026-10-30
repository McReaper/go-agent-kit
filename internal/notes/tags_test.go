package notes_test

import (
	"bytes"
	"testing"
	"time"

	"example.com/notes/internal/notes"
	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPAddAndListByTag(t *testing.T) {
	cs := session(t)
	ctx := t.Context()
	for _, args := range []map[string]any{
		{"title": "Milk", "tags": []string{"home"}},
		{"title": "Report", "tags": []string{"work"}},
	} {
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "add_note", Arguments: args})
		if err != nil || res.IsError {
			t.Fatalf("add_note %v = %+v, %v", args, res, err)
		}
	}
	listed, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "list_notes", Arguments: map[string]any{"tag": "work"}})
	if err != nil || listed.IsError {
		t.Fatalf("list_notes = %+v, %v", listed, err)
	}
	want := map[string]any{"notes": []any{map[string]any{
		"id": "report-20261030t220509", "title": "Report", "tags": []any{"work"},
	}}}
	if diff := cmp.Diff(want, listed.StructuredContent); diff != "" {
		t.Errorf("list_notes (-want +got):\n%s", diff)
	}
}

func TestMCPBadTagIsAToolError(t *testing.T) {
	res, err := session(t).CallTool(t.Context(), &mcp.CallToolParams{
		Name:      "add_note",
		Arguments: map[string]any{"title": "x", "tags": []string{"Work"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatalf("add_note accepted an upper-case tag: %+v", res)
	}
}

func TestCLIAddAndListByTag(t *testing.T) {
	var out bytes.Buffer
	m, err := notes.Open(t.TempDir(), func() time.Time { return time.Date(2026, 10, 30, 22, 5, 9, 0, time.UTC) })
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := m.Close(); err != nil {
			t.Error(err)
		}
	})
	cmds, err := m.Commands(&out)
	if err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) {
		t.Helper()
		for _, c := range cmds {
			if c.Name() == args[0] {
				c.SetArgs(args[1:])
				if err := c.ExecuteContext(t.Context()); err != nil {
					t.Fatal(err)
				}
				return
			}
		}
		t.Fatalf("no command %s", args[0])
	}
	run("add-note", "--title", "Report", "--tags", "work,urgent")
	out.Reset()
	run("list-notes", "--tag", "urgent")
	want := `{"notes":[{"id":"report-20261030t220509","title":"Report","tags":["work","urgent"]}]}` + "\n"
	if diff := cmp.Diff(want, out.String()); diff != "" {
		t.Errorf("list-notes (-want +got):\n%s", diff)
	}
}
