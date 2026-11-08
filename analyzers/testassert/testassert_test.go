package testassert_test

import (
	"testing"

	"github.com/McReaper/rumblestrip-go/analyzers/testassert"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), testassert.New(), "ta")
}
