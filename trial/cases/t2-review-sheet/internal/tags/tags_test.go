package tags_test

import (
	"slices"
	"testing"

	"github.com/McReaper/go-agent-kit/template/internal/tags"
)

func TestNormalizeKeepsTheFirstOrderOnce(t *testing.T) {
	t.Run("case folded", func(t *testing.T) {
		if got := tags.Normalize([]string{"Go", "go", " GO "}); !slices.Equal(got, []string{"go"}) {
			t.Fatalf("Normalize = %v, want [go]", got)
		}
	})
}
