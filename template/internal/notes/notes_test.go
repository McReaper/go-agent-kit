package notes_test

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/McReaper/go-agent-kit/template/internal/notes"
	"github.com/McReaper/go-agent-kit/template/internal/platform/journal"
	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/goleak"
)

var update = flag.Bool("update", false, "rewrite the golden files")

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func clock() time.Time { return time.Date(2026, 10, 30, 22, 5, 9, 0, time.UTC) }

func session(t *testing.T, dir string) *mcp.ClientSession {
	t.Helper()
	m, err := notes.Open(filepath.Join(dir, "notes"), clock)
	if err != nil {
		t.Fatal(err)
	}
	j, err := journal.Open(dir, clock)
	if err != nil {
		t.Fatal(err)
	}
	s := mcp.NewServer(&mcp.Implementation{Name: "notes", Version: "test"}, nil)
	if err := m.Register(s, j); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	st, ct := mcp.NewInMemoryTransports()
	ss, err := s.Connect(ctx, st, nil)
	if err != nil {
		t.Fatal(err)
	}
	cs, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := errors.Join(cs.Close(), ss.Wait(), m.Close(), j.Close()); err != nil {
			t.Error(err)
		}
	})
	return cs
}

func TestToolsMatchTheGoldenContract(t *testing.T) {
	res, err := session(t, t.TempDir()).ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	got, err := json.MarshalIndent(res.Tools, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	golden := filepath.Join("testdata", "tools.golden.json")
	if *update {
		if err := os.WriteFile(golden, append(got, '\n'), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(string(want), string(got)+"\n"); diff != "" {
		t.Errorf("tools changed (-golden +got), rerun with -update if intended:\n%s", diff)
	}
}

func TestAnAddedNoteIsListedAndJournaled(t *testing.T) {
	dir := t.TempDir()
	cs := session(t, dir)
	ctx := context.Background()
	added, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "add_note", Arguments: map[string]any{"title": "Buy milk"}})
	if err != nil || added.IsError {
		t.Fatalf("add_note = %+v, %v", added, err)
	}
	listed, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "list_notes", Arguments: map[string]any{}})
	if err != nil || listed.IsError {
		t.Fatalf("list_notes = %+v, %v", listed, err)
	}
	want := map[string]any{"notes": []any{map[string]any{"id": "buy-milk-20261030t220509", "title": "Buy milk"}}}
	if diff := cmp.Diff(want, listed.StructuredContent); diff != "" {
		t.Errorf("list_notes (-want +got):\n%s", diff)
	}
	data, err := os.ReadFile(filepath.Join(dir, journal.File))
	if err != nil {
		t.Fatal(err)
	}
	if got := string(data); got != `{"time":"2026-10-30T22:05:09Z","op":"add_note","args":{"title":"Buy milk"}}`+"\n" {
		t.Errorf("journal = %q, want add_note alone: list_notes only reads", got)
	}
}

func TestATitleOverTheLimitIsAToolError(t *testing.T) {
	res, err := session(t, t.TempDir()).CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "add_note",
		Arguments: map[string]any{"title": strings.Repeat("a", 81)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatalf("add_note accepted a title over the limit: %+v", res)
	}
}
