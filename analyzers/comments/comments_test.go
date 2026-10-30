package comments_test

import (
	"testing"

	"github.com/McReaper/go-agent-kit/analyzers/comments"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), comments.New(), "example.com/app/internal/notes", "example.com/app/internal/notes/internal/add", "example.com/app/internal/notes/internal/legal", "example.com/app/cmd/app")
}
