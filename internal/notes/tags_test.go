package notes_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestListNotesReturnsOnlyTheNotesWithTheTag(t *testing.T) {
	cs := session(t, t.TempDir())
	ctx := context.Background()
	for _, args := range []map[string]any{
		{"title": "Buy milk", "tags": []string{"Errands"}},
		{"title": "Read book"},
	} {
		if res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "add_note", Arguments: args}); err != nil || res.IsError {
			t.Fatalf("add_note %v = %+v, %v", args, res, err)
		}
	}
	listed, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "list_notes", Arguments: map[string]any{"tag": "errands"}})
	if err != nil || listed.IsError {
		t.Fatalf("list_notes = %+v, %v", listed, err)
	}
	want := map[string]any{"notes": []any{map[string]any{
		"id": "buy-milk-20261030t220509", "title": "Buy milk", "tags": []any{"errands"},
	}}}
	if diff := cmp.Diff(want, listed.StructuredContent); diff != "" {
		t.Errorf("list_notes (-want +got):\n%s", diff)
	}
}
