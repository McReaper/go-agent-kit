package journal_test

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/McReaper/go-agent-kit/template/internal/platform/journal"
	"github.com/google/go-cmp/cmp"
)

var fixedClock = func() time.Time { return time.Date(2026, 10, 31, 9, 30, 0, 0, time.FixedZone("CET", 3600)) }

func entries(t *testing.T, dir string) []journal.Entry {
	t.Helper()
	f, err := os.Open(filepath.Join(dir, journal.File))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})
	var out []journal.Entry
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var e journal.Entry
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			t.Fatalf("line %q: %v", sc.Text(), err)
		}
		out = append(out, e)
	}
	return out
}

func TestTheJournalAppendsEachRunWithItsArgumentsAndFailure(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "new")
	for i, failure := range []error{nil, errors.New("disk full")} {
		j, err := journal.Open(dir, fixedClock)
		if err != nil {
			t.Fatal(err)
		}
		if err := j.Record("add_note", map[string]int{"run": i}, failure); err != nil {
			t.Fatal(err)
		}
		if err := j.Close(); err != nil {
			t.Fatal(err)
		}
	}
	at := time.Date(2026, 10, 31, 8, 30, 0, 0, time.UTC)
	want := []journal.Entry{
		{Time: at, Op: "add_note", Args: json.RawMessage(`{"run":0}`)},
		{Time: at, Op: "add_note", Args: json.RawMessage(`{"run":1}`), Error: "disk full"},
	}
	if diff := cmp.Diff(want, entries(t, dir)); diff != "" {
		t.Errorf("journal (-want +got):\n%s", diff)
	}
}

func TestTheJournalKeepsEveryLineOfConcurrentRuns(t *testing.T) {
	dir := t.TempDir()
	j, err := journal.Open(dir, fixedClock)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := range 50 {
		wg.Go(func() {
			if err := j.Record("add_note", map[string]string{"title": string(rune('a' + i%26))}, nil); err != nil {
				t.Error(err)
			}
		})
	}
	wg.Wait()
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	if n := len(entries(t, dir)); n != 50 {
		t.Errorf("%d entries, want 50", n)
	}
}

func TestArgumentsThatDontEncodeAreAnError(t *testing.T) {
	j, err := journal.Open(t.TempDir(), fixedClock)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := j.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := j.Record("add_note", func() {}, nil); err == nil {
		t.Error("no error for arguments JSON can't encode")
	}
}
