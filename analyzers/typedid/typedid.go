// Package typedid reports identifiers held in a bare string or integer in a
// module's domain (criterion E4): with every id a string, a note id passes
// where a user id was meant and nothing notices.
package typedid

import (
	"go/ast"
	"go/types"
	"regexp"

	"github.com/McReaper/go-agent-kit/analyzers/internal/layout"
	"golang.org/x/tools/go/analysis"
)

// New returns the analyzer.
func New() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "typedid",
		Doc:  "reports ids of a bare string or integer type in a module's domain",
		Run:  layout.Run(run),
	}
}

var idName = regexp.MustCompile(`^(id|ID)$|[a-z0-9](ID|Id)$`)

func run(pass *analysis.Pass) error {
	if !layout.IsDomain(pass.Pkg) {
		return nil
	}
	for _, f := range pass.Files {
		if layout.IsTest(pass.Fset, f) {
			continue
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if fl, ok := n.(*ast.FieldList); ok {
				check(pass, fl)
			}
			return true
		})
	}
	return nil
}

func check(pass *analysis.Pass, fl *ast.FieldList) {
	for _, field := range fl.List {
		b, ok := pass.TypesInfo.TypeOf(field.Type).(*types.Basic)
		if ok && b.Info()&(types.IsString|types.IsInteger) != 0 {
			checkNames(pass, field.Names, b)
		}
	}
}

func checkNames(pass *analysis.Pass, names []*ast.Ident, b *types.Basic) {
	for _, name := range names {
		if idName.MatchString(name.Name) {
			pass.Reportf(name.Pos(), "E4: %s is a bare %s: give ids a type of their own (type NoteID string) "+
				"so ids of different things can't be swapped", name.Name, b.Name())
		}
	}
}
