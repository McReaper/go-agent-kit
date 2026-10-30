// Package testassert reports tests that can only fail on an error or a nil
// (criterion M6): a test whose every failure sits under `if err != nil` or
// `if x == nil` checks that the code ran, not what it did.
//
// A failure is a call to Error, Errorf, Fatal, Fatalf, Fail or FailNow on a
// test's T. Passing the T to another function (a helper, rapid.Check) counts
// as a real check, since its body isn't visible here.
package testassert

import (
	"go/ast"
	"go/token"
	"go/types"
	"slices"
	"strings"

	"github.com/McReaper/go-agent-kit/analyzers/internal/layout"
	"golang.org/x/tools/go/analysis"
)

// New returns the analyzer.
func New() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "testassert",
		Doc:  "reports tests that can only fail on an error or a nil",
		Run:  layout.Run(run),
	}
}

func run(pass *analysis.Pass) error {
	for _, f := range pass.Files {
		if layout.IsTest(pass.Fset, f) {
			checkFile(pass, f)
		}
	}
	return nil
}

func checkFile(pass *analysis.Pass, f *ast.File) {
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if ok && isTest(pass, fn) && !checksSomething(pass, fn.Body) {
			pass.Reportf(fn.Pos(), "M6: %s only fails on an error or a nil: assert what the code returned "+
				"(compare with cmp.Diff, or check a field)", fn.Name.Name)
		}
	}
}

func isFailureMethod(name string) bool {
	switch name {
	case "Error", "Errorf", "Fatal", "Fatalf", "Fail", "FailNow":
		return true
	}
	return false
}

func isTest(pass *analysis.Pass, fn *ast.FuncDecl) bool {
	params := fn.Type.Params.List
	return fn.Recv == nil && fn.Body != nil && strings.HasPrefix(fn.Name.Name, "Test") &&
		len(params) == 1 && len(params[0].Names) <= 1 && isT(pass.TypesInfo.TypeOf(params[0].Type))
}

func checksSomething(pass *analysis.Pass, body *ast.BlockStmt) bool {
	var parents []ast.Node
	checked := false
	ast.Inspect(body, func(n ast.Node) bool {
		switch {
		case checked:
			return false
		case n == nil:
			parents = parents[:len(parents)-1]
			return false
		}
		if call, ok := n.(*ast.CallExpr); ok {
			checked = isFailure(pass, call) && !underNilCheck(pass, parents, call) || passesT(pass, call)
		}
		parents = append(parents, n)
		return true
	})
	return checked
}

func isFailure(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	return ok && isFailureMethod(sel.Sel.Name) && isT(pass.TypesInfo.TypeOf(sel.X))
}

func passesT(pass *analysis.Pass, call *ast.CallExpr) bool {
	for _, a := range call.Args {
		if isT(pass.TypesInfo.TypeOf(a)) {
			return true
		}
	}
	return false
}

func isT(t types.Type) bool {
	if p, ok := t.(*types.Pointer); ok {
		t = p.Elem()
	}
	n, ok := t.(*types.Named)
	return ok && n.Obj().Name() == "T"
}

func underNilCheck(pass *analysis.Pass, parents []ast.Node, call *ast.CallExpr) bool {
	for _, p := range slices.Backward(parents) {
		if s, ok := p.(*ast.IfStmt); ok && call.Pos() >= s.Body.Pos() && call.End() <= s.Body.End() {
			return isNilComparison(pass, s.Cond)
		}
	}
	return false
}

func isNilComparison(pass *analysis.Pass, cond ast.Expr) bool {
	b, ok := cond.(*ast.BinaryExpr)
	if !ok {
		return false
	}
	switch {
	case isNil(pass, b.Y):
		return b.Op == token.NEQ && isError(pass, b.X) || !isError(pass, b.X) && (b.Op == token.EQL || b.Op == token.NEQ)
	case isNil(pass, b.X):
		return b.Op == token.NEQ && isError(pass, b.Y) || !isError(pass, b.Y) && (b.Op == token.EQL || b.Op == token.NEQ)
	}
	return false
}

func isNil(pass *analysis.Pass, e ast.Expr) bool {
	tv, ok := pass.TypesInfo.Types[e]
	return ok && tv.IsNil()
}

func isError(pass *analysis.Pass, e ast.Expr) bool {
	t := pass.TypesInfo.TypeOf(e)
	return t != nil && types.Identical(t, types.Universe.Lookup("error").Type())
}
