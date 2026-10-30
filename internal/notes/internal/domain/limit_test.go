package domain_test

import (
	"strings"
	"testing"

	"example.com/notes/internal/notes/internal/domain"
)

func TestParseLimit(t *testing.T) {
	cases := []struct {
		name string
		in   int
		want int
	}{
		{"the default when none is given", 0, domain.DefaultLimit},
		{"the one given", 3, 3},
		{"the smallest", 1, 1},
		{"the largest", domain.MaxLimit, domain.MaxLimit},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := domain.ParseLimit(c.in)
			if err != nil || got != c.want {
				t.Fatalf("ParseLimit(%d) = %d, %v; want %d, nil", c.in, got, err, c.want)
			}
		})
	}
}

func TestParseLimitRefusesOutOfRange(t *testing.T) {
	for _, n := range []int{-1, domain.MaxLimit + 1} {
		_, err := domain.ParseLimit(n)
		if err == nil || !strings.Contains(err.Error(), "give 1 to 100") {
			t.Errorf("ParseLimit(%d) err = %v, want the range it must be in", n, err)
		}
	}
}
