package domainconv_test

import (
	"testing"

	"github.com/McReaper/go-agent-kit/analyzers/domainconv"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), domainconv.New(), "example.com/app/internal/notes/internal/domain", "example.com/app/internal/notes/internal/edit")
}
