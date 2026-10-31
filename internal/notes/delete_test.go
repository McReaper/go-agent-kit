package notes_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/McReaper/go-agent-kit/template/internal/platform/journal"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestADeletedNoteIsGoneAndJournaled(t *testing.T) {
	dir := t.TempDir()
	cs := session(t, dir)
	ctx := context.Background()
	if res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "add_note", Arguments: map[string]any{"title": "Buy milk"}}); err != nil || res.IsError {
		t.Fatalf("add_note = %+v, %v", res, err)
	}
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "delete_note", Arguments: map[string]any{"id": "buy-milk-20261030t220509"}})
	if err != nil || res.IsError {
		t.Fatalf("delete_note = %+v, %v", res, err)
	}
	listed, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "list_notes", Arguments: map[string]any{}})
	if err != nil || listed.IsError {
		t.Fatalf("list_notes = %+v, %v", listed, err)
	}
	if got := listed.StructuredContent.(map[string]any)["notes"]; got != nil && len(got.([]any)) != 0 {
		t.Errorf("notes after delete = %v, want none", got)
	}
	data, err := os.ReadFile(filepath.Join(dir, journal.File))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"op":"delete_note","args":{"id":"buy-milk-20261030t220509"}`) {
		t.Errorf("journal = %q, want the delete_note run", data)
	}
}

func TestDeletingAnUnknownIDIsAToolErrorThatSaysSo(t *testing.T) {
	res, err := session(t, t.TempDir()).CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "delete_note",
		Arguments: map[string]any{"id": "no-such-note-20261030t220509"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatalf("delete_note of a missing id succeeded: %+v", res)
	}
	if text := res.Content[0].(*mcp.TextContent).Text; !strings.Contains(text, "no note has this id") {
		t.Errorf("error = %q, want it to say no note has this id", text)
	}
}
