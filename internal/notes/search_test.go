package notes_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestSearchNotesFindsTheNotesWhoseTitleOrBodyContainsTheText(t *testing.T) {
	cs := session(t, t.TempDir())
	ctx := context.Background()
	for _, args := range []map[string]any{
		{"title": "Buy milk"},
		{"title": "Read book", "body": "a chapter about DAIRY farms"},
		{"title": "Call mum"},
	} {
		if res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "add_note", Arguments: args}); err != nil || res.IsError {
			t.Fatalf("add_note %v = %+v, %v", args, res, err)
		}
	}
	found, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "search_notes", Arguments: map[string]any{"text": "dairy"}})
	if err != nil || found.IsError {
		t.Fatalf("search_notes = %+v, %v", found, err)
	}
	want := map[string]any{"notes": []any{map[string]any{
		"id": "read-book-20261030t220509", "title": "Read book", "body": "a chapter about DAIRY farms",
	}}}
	if diff := cmp.Diff(want, found.StructuredContent); diff != "" {
		t.Errorf("search_notes (-want +got):\n%s", diff)
	}
}
