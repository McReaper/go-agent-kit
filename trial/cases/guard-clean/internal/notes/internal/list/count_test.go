package list_test

import (
	"testing"

	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/list"
)

func TestCountCountsTheNotes(t *testing.T) {
	if n := list.Count(list.Out{Notes: make([]list.Item, 3)}); n != 3 {
		t.Errorf("Count = %d", n)
	}
}
