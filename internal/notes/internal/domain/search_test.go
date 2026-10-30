package domain_test

import (
	"testing"
	"time"

	"example.com/notes/internal/notes/internal/domain"
)

func TestContainsIgnoresCaseInTheTitleAndTheText(t *testing.T) {
	n := domain.New("Buy Milk", "two litres, OAT", time.Date(2026, 10, 30, 22, 0, 0, 0, time.UTC))
	cases := []struct {
		text string
		want bool
	}{
		{"milk", true},
		{"MILK", true},
		{"litres", true},
		{"oat", true},
		{"Oat", true},
		{"bread", false},
		{"", true},
	}
	for _, c := range cases {
		if got := n.Contains(c.text); got != c.want {
			t.Errorf("Contains(%q) = %v, want %v", c.text, got, c.want)
		}
	}
}

func TestCheckLimitDefaultsAndBounds(t *testing.T) {
	cases := []struct {
		name  string
		limit int
		want  int
		err   bool
	}{
		{"the default when none is given", 0, domain.DefaultLimit, false},
		{"the one given", 7, 7, false},
		{"the smallest", 1, 1, false},
		{"the largest", domain.MaxLimit, domain.MaxLimit, false},
		{"negative", -1, 0, true},
		{"over the largest", domain.MaxLimit + 1, 0, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := domain.CheckLimit(c.limit)
			if (err != nil) != c.err {
				t.Fatalf("CheckLimit(%d) err = %v, want error: %v", c.limit, err, c.err)
			}
			if got != c.want {
				t.Errorf("CheckLimit(%d) = %d, want %d", c.limit, got, c.want)
			}
		})
	}
}
