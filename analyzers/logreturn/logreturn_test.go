package logreturn_test

import (
	"testing"

	"github.com/McReaper/rumblestrip-go/analyzers/logreturn"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), logreturn.New(), "lr")
}
