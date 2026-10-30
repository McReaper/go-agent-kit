// Package ctxroot reports new root contexts outside main (criterion M11):
// context.Background or context.TODO deep in the code cuts the caller's
// cancellation and deadline. Tests use t.Context(), which usetesting checks.
package ctxroot

import (
	"go/ast"
	"go/types"

	"github.com/McReaper/go-agent-kit/analyzers/internal/layout"
	"golang.org/x/tools/go/analysis"
)

// New returns the analyzer.
func New() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "ctxroot",
		Doc:  "reports context.Background and context.TODO outside main packages",
		Run:  layout.Run(run),
	}
}

func run(pass *analysis.Pass) error {
	if layout.IsMain(pass.Pkg) {
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
	if !ok {
		return
	}
	if name, ok := rootContext(pass, call); ok {
		pass.Reportf(call.Pos(), "M11: context.%s starts a new root: take the ctx your caller passes in, "+
			"so its cancellation and deadline reach this code", name)
	}
}

func rootContext(pass *analysis.Pass, call *ast.CallExpr) (string, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}
	fn, ok := pass.TypesInfo.Uses[sel.Sel].(*types.Func)
	if !ok || fn.Pkg() == nil || fn.Pkg().Path() != "context" {
		return "", false
	}
	return fn.Name(), fn.Name() == "Background" || fn.Name() == "TODO"
}
