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

	"example.com/notes/internal/notes"
	"github.com/google/go-cmp/cmp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/goleak"
)

var update = flag.Bool("update", false, "rewrite the golden files")

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func session(t *testing.T) *mcp.ClientSession {
	t.Helper()
	m, err := notes.Open(t.TempDir(), func() time.Time { return time.Date(2026, 10, 30, 22, 5, 9, 0, time.UTC) })
	if err != nil {
		t.Fatal(err)
	}
	s := mcp.NewServer(&mcp.Implementation{Name: "notes", Version: "test"}, nil)
	if err := m.Register(s); err != nil {
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
		if err := errors.Join(cs.Close(), ss.Wait(), m.Close()); err != nil {
			t.Error(err)
		}
	})
	return cs
}

func TestToolsMatchTheGoldenContract(t *testing.T) {
	res, err := session(t).ListTools(context.Background(), nil)
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

func TestAnAddedNoteIsListed(t *testing.T) {
	cs := session(t)
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
}

func TestATitleOverTheLimitIsAToolError(t *testing.T) {
	res, err := session(t).CallTool(context.Background(), &mcp.CallToolParams{
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
