// Package goroutine reports go statements outside platform code (criterion
// M11). A goroutine started in a slice has no owner: nothing waits for it,
// cancels it or sees its error. The platform starts them, through
// errgroup.Group or sync.WaitGroup.Go, and tests may start their own.
package goroutine

import (
	"go/ast"

	"github.com/McReaper/go-agent-kit/analyzers/internal/layout"
	"golang.org/x/tools/go/analysis"
)

// New returns the analyzer.
func New() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "goroutine",
		Doc:  "reports go statements outside platform code and tests",
		Run:  layout.Run(run),
	}
}

func run(pass *analysis.Pass) error {
	if layout.IsPlatform(pass.Pkg) {
		return nil
	}
	for _, f := range pass.Files {
		if layout.IsTest(pass.Fset, f) {
			continue
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if g, ok := n.(*ast.GoStmt); ok {
				pass.Reportf(g.Pos(), "M11: a goroutine started here has no owner: start it in platform code, "+
					"through errgroup.Group or sync.WaitGroup.Go, so something waits for it and sees its error")
			}
			return true
		})
	}
	return nil
}
