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

func TestEditNoteChangesTheNoteListedUnderTheSameID(t *testing.T) {
	cs := session(t)
	ctx := context.Background()
	if _, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "add_note", Arguments: map[string]any{"title": "Buy milk"}}); err != nil {
		t.Fatal(err)
	}
	edited, err := cs.CallTool(ctx, &mcp.CallToolParams{
		Name:      "edit_note",
		Arguments: map[string]any{"id": "buy-milk-20261030t220509", "title": "Buy oat milk"},
	})
	if err != nil || edited.IsError {
		t.Fatalf("edit_note = %+v, %v", edited, err)
	}
	listed, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "list_notes", Arguments: map[string]any{}})
	if err != nil || listed.IsError {
		t.Fatalf("list_notes = %+v, %v", listed, err)
	}
	want := map[string]any{"notes": []any{map[string]any{"id": "buy-milk-20261030t220509", "title": "Buy oat milk"}}}
	if diff := cmp.Diff(want, listed.StructuredContent); diff != "" {
		t.Errorf("list_notes after edit_note (-want +got):\n%s", diff)
	}
}

func TestEditNoteOfAnUnknownIDIsAToolError(t *testing.T) {
	res, err := session(t).CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "edit_note",
		Arguments: map[string]any{"id": "nope-20261030t220104", "title": "x"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatalf("edit_note accepted an unknown id: %+v", res)
	}
}

func TestEditNoteCommandPrintsTheEditedNote(t *testing.T) {
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
	byName := map[string]func(args ...string) error{}
	for _, c := range cmds {
		byName[c.Name()] = func(args ...string) error {
			c.SetArgs(args)
			return c.ExecuteContext(context.Background())
		}
	}
	if err := byName["add-note"]("--title", "Buy milk", "--body", "two litres"); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := byName["edit-note"]("--id", "buy-milk-20261030t220509", "--clear-body"); err != nil {
		t.Fatal(err)
	}
	var got map[string]string
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("printed %q: %v", out.String(), err)
	}
	want := map[string]string{"id": "buy-milk-20261030t220509", "title": "Buy milk"}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("printed (-want +got):\n%s", diff)
	}
}
