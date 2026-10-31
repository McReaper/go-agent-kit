package racecase_test

import (
	"sync"
	"testing"

	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/racecase"
)

func TestIncFromTwoGoroutines(t *testing.T) {
	var c racecase.Counter
	var wg sync.WaitGroup
	wg.Go(c.Inc)
	wg.Go(c.Inc)
	wg.Wait()
	if c.N() != 2 {
		t.Errorf("N = %d, want 2", c.N())
	}
}
