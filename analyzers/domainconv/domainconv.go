// Package domainconv reports a domain type built by conversion outside the
// domain (criterion E4): domain.NoteID(s) skips the domain's Parse function,
// so a value no rule checked enters the model. Inside the domain, and in
// tests, conversions are free.
package domainconv

import (
	"go/ast"
	"go/types"

	"github.com/McReaper/go-agent-kit/analyzers/internal/layout"
	"golang.org/x/tools/go/analysis"
)

// New returns the analyzer.
func New() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "domainconv",
		Doc:  "reports a domain type built by conversion outside the domain",
		Run:  layout.Run(run),
	}
}

func run(pass *analysis.Pass) error {
	if layout.IsDomain(pass.Pkg) {
		return nil
	}
	for _, f := range pass.Files {
		if !layout.IsTest(pass.Fset, f) {
			ast.Inspect(f, func(n ast.Node) bool {
				check(pass, n)
				return true
			})
		}
	}
	return nil
}

func check(pass *analysis.Pass, n ast.Node) {
	call, ok := n.(*ast.CallExpr)
	if !ok || len(call.Args) != 1 {
		return
	}
	if named, ok := domainType(pass, call.Fun); ok {
		pass.Reportf(call.Pos(), "E4: %s is built by conversion outside the domain: call the domain's Parse%s "+
			"so the value is checked (add it to the domain if it has none)", named.Obj().Name(), named.Obj().Name())
	}
}

func domainType(pass *analysis.Pass, fun ast.Expr) (*types.Named, bool) {
	tv, ok := pass.TypesInfo.Types[fun]
	if !ok || !tv.IsType() {
		return nil, false
	}
	named, ok := tv.Type.(*types.Named)
	if !ok || named.Obj().Pkg() == nil || !layout.IsDomain(named.Obj().Pkg()) {
		return nil, false
	}
	_, basic := named.Underlying().(*types.Basic)
	return named, basic
}
