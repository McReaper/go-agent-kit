package nesting_test

import (
	"testing"

	"github.com/McReaper/rumblestrip-go/analyzers/nesting"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), nesting.New(), "n")
}
