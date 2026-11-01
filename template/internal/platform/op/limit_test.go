package op_test

import (
	"strings"
	"testing"

	"github.com/McReaper/go-agent-kit/template/internal/platform/op"
)

func TestALimitIsTheDefaultOrOneGivenWithinTheRange(t *testing.T) {
	cases := []struct {
		given *int
		want  int
	}{
		{nil, 20},
		{new(1), 1},
		{new(100), 100},
	}
	for _, c := range cases {
		if got, err := op.Limit(c.given, 20, 100); err != nil || got != c.want {
			t.Errorf("Limit(%v) = %d, %v; want %d", c.given, got, err, c.want)
		}
	}
	for _, n := range []int{0, -1, 101} {
		if _, err := op.Limit(new(n), 20, 100); err == nil || !strings.Contains(err.Error(), "give 1 to 100") {
			t.Errorf("Limit(%d): err = %v, want the range named", n, err)
		}
	}
}
