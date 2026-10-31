package notes_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestSearchNotesOverMCPAndCLI(t *testing.T) {
	cs := session(t, t.TempDir())
	ctx := context.Background()
	for _, args := range []map[string]any{
		{"title": "Buy milk"},
		{"title": "Groceries", "body": "Two litres of MILK"},
		{"title": "Bread"},
	} {
		if res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "add_note", Arguments: args}); err != nil || res.IsError {
			t.Fatalf("add_note %v = %+v, %v", args, res, err)
		}
	}
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "search_notes", Arguments: map[string]any{"text": "milk"}})
	if err != nil || res.IsError {
		t.Fatalf("search_notes = %+v, %v", res, err)
	}
	want := map[string]any{"notes": []any{
		map[string]any{"id": "buy-milk-20261030t220509", "title": "Buy milk"},
		map[string]any{"id": "groceries-20261030t220509", "title": "Groceries"},
	}}
	if diff := cmp.Diff(want, res.StructuredContent); diff != "" {
		t.Errorf("search_notes (-want +got):\n%s", diff)
	}
}
