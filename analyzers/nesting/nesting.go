// Package nesting reports control flow nested deeper than two levels
// (criterion M1). Every control statement counts: if, for, range, switch,
// type switch and select; an else-if stays at its if's level. Test files
// are exempt, like the kit's other size rules.
package nesting

import (
	"go/ast"

	"github.com/McReaper/go-agent-kit/analyzers/internal/layout"
	"golang.org/x/tools/go/analysis"
)

// Max is the deepest nesting allowed.
const Max = 2

// New returns the analyzer.
func New() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "nesting",
		Doc:  "reports control flow nested more than two levels deep",
		Run:  layout.Run(run),
	}
}

func run(pass *analysis.Pass) error {
	c := checker{pass: pass}
	for _, f := range pass.Files {
		if !layout.IsTest(pass.Fset, f) {
			c.file(f)
		}
	}
	return nil
}

func (c checker) file(f *ast.File) {
	for _, d := range f.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok && fn.Body != nil {
			c.children(fn.Body, 0)
		}
	}
}

type checker struct {
	pass *analysis.Pass
}

func (c checker) children(n ast.Node, depth int) {
	ast.Inspect(n, func(m ast.Node) bool {
		if m == n {
			return true
		}
		switch s := m.(type) {
		case *ast.IfStmt:
			c.ifStmt(s, depth)
			return false
		case *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt, *ast.TypeSwitchStmt, *ast.SelectStmt:
			c.children(m, c.enter(m, depth))
			return false
		}
		return true
	})
}

func (c checker) ifStmt(s *ast.IfStmt, depth int) {
	inner := c.enter(s, depth)
	if s.Init != nil {
		c.children(s.Init, depth)
	}
	c.children(s.Cond, depth)
	c.children(s.Body, inner)
	switch e := s.Else.(type) {
	case *ast.IfStmt:
		c.ifStmt(e, depth)
	case *ast.BlockStmt:
		c.children(e, inner)
	}
}

func (c checker) enter(n ast.Node, depth int) int {
	if depth+1 == Max+1 {
		c.pass.Reportf(n.Pos(), "M1: control flow nested %d deep, the limit is %d: move the inner block into a function named for what it does", depth+1, Max)
	}
	return depth + 1
}
