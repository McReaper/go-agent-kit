// Package logreturn reports an error both logged and returned in the same
// block (criterion M7): the caller gets it back and logs it again, and one
// failure shows up as several. An error is logged once, where it's handled.
package logreturn

import (
	"go/ast"
	"go/types"

	"github.com/McReaper/go-agent-kit/analyzers/internal/layout"
	"golang.org/x/tools/go/analysis"
)

// New returns the analyzer.
func New() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "logreturn",
		Doc:  "reports an error logged with slog and returned in the same block",
		Run:  layout.Run(run),
	}
}

func run(pass *analysis.Pass) error {
	for _, f := range pass.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			if b, ok := n.(*ast.BlockStmt); ok {
				checkBlock(pass, b)
			}
			return true
		})
	}
	return nil
}

func checkBlock(pass *analysis.Pass, b *ast.BlockStmt) {
	returned := returnedErrors(pass, b)
	for _, s := range b.List {
		if call, ok := logCall(pass, s); ok {
			reportIfReturned(pass, call, returned)
		}
	}
}

func returnedErrors(pass *analysis.Pass, b *ast.BlockStmt) map[types.Object]bool {
	returned := map[types.Object]bool{}
	for _, s := range b.List {
		if r, ok := s.(*ast.ReturnStmt); ok {
			errorsIn(pass, r, returned)
		}
	}
	return returned
}

func reportIfReturned(pass *analysis.Pass, call *ast.CallExpr, returned map[types.Object]bool) {
	logged := map[types.Object]bool{}
	errorsIn(pass, call, logged)
	for obj := range logged {
		if returned[obj] {
			pass.Reportf(call.Pos(), "M7: %s is logged and returned: the caller logs it again; "+
				"return it wrapped with context, or log it here and handle it", obj.Name())
			return
		}
	}
}

func logCall(pass *analysis.Pass, s ast.Stmt) (*ast.CallExpr, bool) {
	es, ok := s.(*ast.ExprStmt)
	if !ok {
		return nil, false
	}
	call, ok := es.X.(*ast.CallExpr)
	if !ok {
		return nil, false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil, false
	}
	fn, ok := pass.TypesInfo.Uses[sel.Sel].(*types.Func)
	return call, ok && fn.Pkg() != nil && fn.Pkg().Path() == "log/slog"
}

func errorsIn(pass *analysis.Pass, n ast.Node, into map[types.Object]bool) {
	errType := types.Universe.Lookup("error").Type()
	ast.Inspect(n, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			if obj := pass.TypesInfo.Uses[id]; obj != nil && types.Identical(obj.Type(), errType) {
				into[obj] = true
			}
		}
		return true
	})
}
