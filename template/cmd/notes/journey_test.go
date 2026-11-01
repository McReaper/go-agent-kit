package main_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type listed struct {
	Notes []struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	} `json:"notes"`
}

func TestAUserAddsANoteFromTheCommandLineThenFindsItInTheList(t *testing.T) {
	bin, dataDir := build(t), t.TempDir()
	r := run(t, bin, dataDir, "add-note", "--title=Groceries", "--body=eggs, milk")
	if r.code != 0 {
		t.Fatalf("add-note exited %d:\n%s", r.code, r.stderr)
	}
	var added struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(r.stdout), &added); err != nil || !strings.HasPrefix(added.ID, "groceries-") {
		t.Fatalf("add-note printed %q (%v), want the new note's id", r.stdout, err)
	}
	r = run(t, bin, dataDir, "list-notes")
	var got listed
	if err := json.Unmarshal([]byte(r.stdout), &got); err != nil {
		t.Fatalf("list-notes printed %q: %v", r.stdout, err)
	}
	want := listed{Notes: []struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}{{added.ID, "Groceries"}}}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("list-notes (-want +got):\n%s", diff)
	}
}

func TestARefusedNoteExitsWithOneAndABadFlagWithTwo(t *testing.T) {
	bin := build(t)
	r := run(t, bin, t.TempDir(), "add-note", "--title="+strings.Repeat("x", 81))
	if r.code != 1 || !strings.Contains(r.stderr, "the limit is 80") {
		t.Errorf("exit %d, stderr %q: want 1 and the title's limit", r.code, r.stderr)
	}
	r = run(t, bin, t.TempDir(), "add-note", "--no-such-flag")
	if r.code != 2 {
		t.Errorf("an unknown flag exited %d, want 2", r.code)
	}
}

func TestAnAgentAddsANoteThroughMCPThenFindsItInTheList(t *testing.T) {
	s := connect(t, build(t), t.TempDir())
	add, err := s.CallTool(t.Context(), &mcp.CallToolParams{Name: "add_note", Arguments: map[string]any{"title": "Groceries"}})
	if err != nil || add.IsError {
		t.Fatalf("add_note: %v, %+v", err, add)
	}
	list, err := s.CallTool(t.Context(), &mcp.CallToolParams{Name: "list_notes", Arguments: map[string]any{}})
	if err != nil || list.IsError {
		t.Fatalf("list_notes: %v, %+v", err, list)
	}
	data, err := json.Marshal(list.StructuredContent)
	if err != nil {
		t.Fatal(err)
	}
	var got listed
	if err := json.Unmarshal(data, &got); err != nil || len(got.Notes) != 1 || got.Notes[0].Title != "Groceries" {
		t.Errorf("list_notes returned %s (%v), want the one note added", data, err)
	}
}
