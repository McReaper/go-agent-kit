package ctxroot_test

import (
	"testing"

	"github.com/McReaper/rumblestrip-go/analyzers/ctxroot"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), ctxroot.New(), "lib", "app")
}
