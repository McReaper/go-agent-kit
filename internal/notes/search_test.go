package notes_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestSearchNotesFindsTheMatchingNoteOverMCP(t *testing.T) {
	cs := session(t)
	ctx := context.Background()
	for _, title := range []string{"Buy milk", "Bread"} {
		if _, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "add_note", Arguments: map[string]any{"title": title}}); err != nil {
			t.Fatal(err)
		}
	}
	found, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "search_notes", Arguments: map[string]any{"text": "MILK"}})
	if err != nil || found.IsError {
		t.Fatalf("search_notes = %+v, %v", found, err)
	}
	want := map[string]any{"notes": []any{map[string]any{"id": "buy-milk-20261030t220509", "title": "Buy milk"}}}
	if diff := cmp.Diff(want, found.StructuredContent); diff != "" {
		t.Errorf("search_notes (-want +got):\n%s", diff)
	}
}

func TestSearchNotesWithoutTextIsAToolError(t *testing.T) {
	res, err := session(t).CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "search_notes",
		Arguments: map[string]any{"text": "  "},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatalf("search_notes accepted an empty text: %+v", res)
	}
}
