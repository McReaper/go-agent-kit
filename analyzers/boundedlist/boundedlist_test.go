package boundedlist_test

import (
	"testing"

	"github.com/McReaper/go-agent-kit/analyzers/boundedlist"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), boundedlist.New(),
		"example.com/app/internal/notes/internal/search", "example.com/app/internal/notes/internal/list",
		"example.com/app/internal/notes/internal/export", "example.com/app/internal/notes/internal/add",
		"example.com/app/internal/notes/internal/infra")
}
