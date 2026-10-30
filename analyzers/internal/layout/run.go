package layout

import "golang.org/x/tools/go/analysis"

// Run adapts a check that has no result to analysis.Analyzer.Run.
//
// WHY: go/analysis requires Run to return a nil result when the analyzer
// declares no ResultType; this adapter is the one place that does, so the
// analyzers themselves never return (nil, nil), which nilnil forbids.
func Run(check func(*analysis.Pass) error) func(*analysis.Pass) (any, error) {
	return func(pass *analysis.Pass) (any, error) {
		return nil, check(pass)
	}
}
