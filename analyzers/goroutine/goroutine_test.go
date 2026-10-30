package goroutine_test

import (
	"testing"

	"github.com/McReaper/go-agent-kit/analyzers/goroutine"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), goroutine.New(), "example.com/app/internal/notes/internal/add", "example.com/app/internal/platform/run")
}
