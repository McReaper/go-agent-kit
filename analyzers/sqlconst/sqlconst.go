// Package sqlconst reports SQL queries built at run time (criterion S2):
// the query given to database/sql must be a constant, its values passed
// as arguments, so no input can change what the query does. gosec's G201
// and G202 miss quoted values and every *Context method.
package sqlconst

import (
	"go/ast"
	"go/types"
	"strings"

	"github.com/McReaper/go-agent-kit/analyzers/internal/layout"
	"golang.org/x/tools/go/analysis"
)

// New returns the analyzer.
func New() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "sqlconst",
		Doc:  "reports database/sql queries that are not constants",
		Run:  layout.Run(run),
	}
}

func run(pass *analysis.Pass) error {
	for _, f := range pass.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok {
				check(pass, call)
			}
			return true
		})
	}
	return nil
}

func check(pass *analysis.Pass, call *ast.CallExpr) {
	i, ok := queryArg(pass, call)
	if !ok || i >= len(call.Args) {
		return
	}
	if tv := pass.TypesInfo.Types[call.Args[i]]; tv.Value == nil {
		pass.Reportf(call.Args[i].Pos(), "S2: this SQL query is built at run time: write it as a constant "+
			"and pass the values as arguments (? placeholders), so no input can change the query")
	}
}

func queryArg(pass *analysis.Pass, call *ast.CallExpr) (int, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return 0, false
	}
	fn, ok := pass.TypesInfo.Uses[sel.Sel].(*types.Func)
	if !ok || fn.Pkg() == nil || fn.Pkg().Path() != "database/sql" || !takesQuery(fn.Name()) {
		return 0, false
	}
	if strings.HasSuffix(fn.Name(), "Context") {
		return 1, true
	}
	return 0, true
}

func takesQuery(name string) bool {
	switch strings.TrimSuffix(name, "Context") {
	case "Query", "QueryRow", "Exec", "Prepare":
		return true
	}
	return false
}
