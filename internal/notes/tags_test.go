package notes_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestTaggedNotesOverMCP(t *testing.T) {
	cs := session(t, t.TempDir())
	ctx := context.Background()
	for _, args := range []map[string]any{
		{"title": "Ship it", "tags": []string{"Go", "go", "Release"}},
		{"title": "Buy milk"},
	} {
		if res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "add_note", Arguments: args}); err != nil || res.IsError {
			t.Fatalf("add_note %v = %+v, %v", args, res, err)
		}
	}
	all, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "list_notes", Arguments: map[string]any{}})
	if err != nil || all.IsError {
		t.Fatalf("list_notes = %+v, %v", all, err)
	}
	wantAll := map[string]any{"notes": []any{
		map[string]any{"id": "buy-milk-20261030t220509", "title": "Buy milk"},
		map[string]any{"id": "ship-it-20261030t220509", "title": "Ship it", "tags": []any{"go", "release"}},
	}}
	if diff := cmp.Diff(wantAll, all.StructuredContent); diff != "" {
		t.Errorf("list_notes (-want +got):\n%s", diff)
	}
	byTag, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "list_notes", Arguments: map[string]any{"tag": "GO"}})
	if err != nil || byTag.IsError {
		t.Fatalf("list_notes by tag = %+v, %v", byTag, err)
	}
	wantTagged := map[string]any{"notes": []any{
		map[string]any{"id": "ship-it-20261030t220509", "title": "Ship it", "tags": []any{"go", "release"}},
	}}
	if diff := cmp.Diff(wantTagged, byTag.StructuredContent); diff != "" {
		t.Errorf("list_notes tag GO (-want +got):\n%s", diff)
	}
}

func TestAnEmptyTagIsAToolError(t *testing.T) {
	res, err := session(t, t.TempDir()).CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "add_note",
		Arguments: map[string]any{"title": "x", "tags": []string{"  "}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatalf("add_note accepted an empty tag: %+v", res)
	}
}
