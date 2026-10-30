package notes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"example.com/notes/internal/notes"
	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func callTool(t *testing.T, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	res, err := session(t).CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestArchiveNoteHidesItFromListAndSearchUntilAsked(t *testing.T) {
	cs := session(t)
	ctx := context.Background()
	call := func(name string, args map[string]any) *mcp.CallToolResult {
		t.Helper()
		res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: args})
		if err != nil || res.IsError {
			t.Fatalf("%s = %+v, %v", name, res, err)
		}
		return res
	}
	call("add_note", map[string]any{"title": "Buy milk"})
	call("add_note", map[string]any{"title": "Bread"})
	call("archive_note", map[string]any{"id": "buy-milk-20261030t220509"})

	want := func(label string, res *mcp.CallToolResult, notes []any) {
		t.Helper()
		if diff := cmp.Diff(map[string]any{"notes": notes}, res.StructuredContent); diff != "" {
			t.Errorf("%s (-want +got):\n%s", label, diff)
		}
	}
	bread := map[string]any{"id": "bread-20261030t220509", "title": "Bread"}
	milk := map[string]any{"id": "buy-milk-20261030t220509", "title": "Buy milk"}

	want("list_notes after archive_note", call("list_notes", map[string]any{}), []any{bread})
	want("search_notes after archive_note", call("search_notes", map[string]any{"text": "milk"}), []any{})
	want("list_notes with include_archived", call("list_notes", map[string]any{"include_archived": true}), []any{bread, milk})
	want("search_notes with include_archived", call("search_notes", map[string]any{"text": "milk", "include_archived": true}), []any{milk})

	call("unarchive_note", map[string]any{"id": "buy-milk-20261030t220509"})
	want("list_notes after unarchive_note", call("list_notes", map[string]any{}), []any{bread, milk})
}

func TestArchiveNoteOfAnUnknownIDIsAToolError(t *testing.T) {
	if res := callTool(t, "archive_note", map[string]any{"id": "nope-20261030t220104"}); !res.IsError {
		t.Fatalf("archive_note accepted an unknown id: %+v", res)
	}
}

func TestArchiveNoteWithoutIDIsAToolError(t *testing.T) {
	if res := callTool(t, "unarchive_note", map[string]any{}); !res.IsError {
		t.Fatalf("unarchive_note accepted no id: %+v", res)
	}
}

func TestArchiveNoteCommandPrintsTheNoteAndHidesIt(t *testing.T) {
	m, err := notes.Open(t.TempDir(), func() time.Time { return time.Date(2026, 10, 30, 22, 5, 9, 0, time.UTC) })
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := m.Close(); err != nil {
			t.Error(err)
		}
	}()
	var out bytes.Buffer
	cmds, err := m.Commands(&out)
	if err != nil {
		t.Fatal(err)
	}
	run := map[string]func(args ...string) error{}
	for _, c := range cmds {
		run[c.Name()] = func(args ...string) error {
			c.SetArgs(args)
			return c.ExecuteContext(context.Background())
		}
	}
	if err := run["add-note"]("--title", "Buy milk"); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := run["archive-note"]("--id", "buy-milk-20261030t220509"); err != nil {
		t.Fatal(err)
	}
	var got map[string]string
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("printed %q: %v", out.String(), err)
	}
	if diff := cmp.Diff(map[string]string{"id": "buy-milk-20261030t220509", "title": "Buy milk"}, got); diff != "" {
		t.Errorf("archive-note printed (-want +got):\n%s", diff)
	}
	out.Reset()
	if err := run["list-notes"](); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(`{"notes":[]}`, string(bytes.TrimSpace(out.Bytes()))); diff != "" {
		t.Errorf("list-notes after archive-note (-want +got):\n%s", diff)
	}
}
