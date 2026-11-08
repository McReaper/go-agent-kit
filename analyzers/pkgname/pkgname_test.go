package pkgname_test

import (
	"testing"

	"github.com/McReaper/rumblestrip-go/analyzers/pkgname"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), pkgname.New(), "utils", "notes")
}
