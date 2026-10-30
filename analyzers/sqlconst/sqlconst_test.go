package sqlconst_test

import (
	"testing"

	"github.com/McReaper/go-agent-kit/analyzers/sqlconst"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), sqlconst.New(), "q")
}
