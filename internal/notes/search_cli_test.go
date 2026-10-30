package notes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"example.com/notes/internal/notes"
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/cobra"
)

func TestSearchNotesCommandPrintsTheNewestMatchesFirst(t *testing.T) {
	var now time.Time
	m, err := notes.Open(t.TempDir(), func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := m.Close(); err != nil {
			t.Error(err)
		}
	}()
	var out bytes.Buffer
	cmds, err := m.Commands(&out)
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]*cobra.Command{}
	for _, c := range cmds {
		byName[c.Name()] = c
	}
	ctx := context.Background()
	run := func(name string, args ...string) {
		t.Helper()
		c := byName[name]
		c.SetArgs(args)
		c.SetContext(ctx)
		if err := c.Execute(); err != nil {
			t.Fatalf("%s %v: %v", name, args, err)
		}
	}
	for minute, title := range []string{"Buy milk", "Oat milk", "Bread"} {
		now = time.Date(2026, 10, 30, 22, minute+1, 0, 0, time.UTC)
		run("add-note", "--title", title)
	}
	out.Reset()
	run("search-notes", "--text", "MILK")

	var got struct {
		Notes []map[string]string `json:"notes"`
	}
	if err := json.NewDecoder(strings.NewReader(out.String())).Decode(&got); err != nil {
		t.Fatalf("output %q: %v", out.String(), err)
	}
	want := []map[string]string{
		{"id": "oat-milk-20261030t220200", "title": "Oat milk"},
		{"id": "buy-milk-20261030t220100", "title": "Buy milk"},
	}
	if diff := cmp.Diff(want, got.Notes); diff != "" {
		t.Errorf("search-notes (-want +got):\n%s", diff)
	}
}
