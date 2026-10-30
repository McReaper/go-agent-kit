package notes_test

import (
	"bytes"
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

func openModule(t *testing.T) *notes.Module {
	t.Helper()
	m, err := notes.Open(t.TempDir(), func() time.Time { return time.Date(2026, 10, 30, 22, 5, 9, 0, time.UTC) })
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := m.Close(); err != nil {
			t.Error(err)
		}
	})
	return m
}

func session(t *testing.T) *mcp.ClientSession {
	t.Helper()
	m := openModule(t)
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
		if err := errors.Join(cs.Close(), ss.Wait()); err != nil {
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

func TestADeletedNoteIsGone(t *testing.T) {
	cs := session(t)
	ctx := context.Background()
	added, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "add_note", Arguments: map[string]any{"title": "Buy milk"}})
	if err != nil || added.IsError {
		t.Fatalf("add_note = %+v, %v", added, err)
	}
	addedOut, ok := added.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("add_note content = %#v", added.StructuredContent)
	}
	deleted, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "delete_note", Arguments: map[string]any{"id": addedOut["id"]}})
	if err != nil || deleted.IsError {
		t.Fatalf("delete_note = %+v, %v", deleted, err)
	}
	listed, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "list_notes", Arguments: map[string]any{}})
	if err != nil || listed.IsError {
		t.Fatalf("list_notes = %+v, %v", listed, err)
	}
	want := map[string]any{"notes": []any{}}
	if diff := cmp.Diff(want, listed.StructuredContent); diff != "" {
		t.Errorf("list_notes after delete_note (-want +got):\n%s", diff)
	}
}

func TestDeleteNoteRefusesAnIDWithNoNoteAsAToolError(t *testing.T) {
	res, err := session(t).CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "delete_note",
		Arguments: map[string]any{"id": "no-such-note-20261030t220509"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsError {
		t.Fatalf("delete_note accepted an id with no note: %+v", res)
	}
}

func TestTheCommandsAddAndDeleteANote(t *testing.T) {
	var out bytes.Buffer
	cmds, err := openModule(t).Commands(&out)
	if err != nil {
		t.Fatal(err)
	}
	run := func(name string, args ...string) error {
		t.Helper()
		for _, c := range cmds {
			if c.Name() == name {
				c.SetArgs(args)
				return c.ExecuteContext(context.Background())
			}
		}
		t.Fatalf("no command %s", name)
		return nil
	}
	if err := run("add-note", "--title", "Buy milk"); err != nil {
		t.Fatal(err)
	}
	var added struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(out.Bytes(), &added); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := run("delete-note", "--id", added.ID); err != nil {
		t.Fatal(err)
	}
	if want := `{"id":"` + added.ID + `"}` + "\n"; out.String() != want {
		t.Errorf("delete-note printed %q, want %q", out.String(), want)
	}
	if err := run("delete-note", "--id", added.ID); err == nil {
		t.Error("deleting the same note twice succeeded")
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
