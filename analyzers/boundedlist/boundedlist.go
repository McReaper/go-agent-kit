// Package boundedlist reports an operation that returns a list without a
// limit (criterion S8): a caller, or an agent in a loop, gets every row,
// however many there are.
package boundedlist

import (
	"go/types"
	"reflect"
	"strings"

	"github.com/McReaper/go-agent-kit/analyzers/internal/layout"
	"golang.org/x/tools/go/analysis"
)

// New returns the analyzer.
func New() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "boundedlist",
		Doc:  "reports a slice whose Out holds a list while its Args has no limit",
		Run:  layout.Run(run),
	}
}

// WHY: a slice is one operation: its Args are what the caller asks, its Out
// what it gets back (the template's convention, which archtest's layout
// enforces); a list in Out is bounded only by a limit in Args.
func run(pass *analysis.Pass) error {
	if !layout.IsSlice(pass.Pkg) {
		return nil
	}
	out, ok := structNamed(pass.Pkg, "Out")
	if !ok {
		return nil
	}
	list := listField(out)
	if list == nil {
		return nil
	}
	if args, ok := structNamed(pass.Pkg, "Args"); ok && hasLimit(args) {
		return nil
	}
	pass.Reportf(list.Pos(), "S8: Out.%s is a list, and Args has no limit: add Limit *int `json:\"limit,omitempty\"` "+
		"with a default and a maximum, so no call returns every row", list.Name())
	return nil
}

func structNamed(pkg *types.Package, name string) (*types.Struct, bool) {
	tn, ok := pkg.Scope().Lookup(name).(*types.TypeName)
	if !ok {
		return nil, false
	}
	s, ok := tn.Type().Underlying().(*types.Struct)
	return s, ok
}

func listField(s *types.Struct) *types.Var {
	for f := range s.Fields() {
		switch f.Type().Underlying().(type) {
		case *types.Slice, *types.Map:
			return f
		}
	}
	return nil
}

func hasLimit(s *types.Struct) bool {
	for i := range s.NumFields() {
		name, _, _ := strings.Cut(reflect.StructTag(s.Tag(i)).Get("json"), ",")
		if name == "limit" && isInteger(s.Field(i).Type()) {
			return true
		}
	}
	return false
}

func isInteger(t types.Type) bool {
	if p, ok := t.Underlying().(*types.Pointer); ok {
		t = p.Elem()
	}
	b, ok := t.Underlying().(*types.Basic)
	return ok && b.Info()&types.IsInteger != 0
}
