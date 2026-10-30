// Package comments reports comments that say what the code does rather
// than why (criterion M2). A comment is allowed when it:
//
//   - starts with "WHY:", anywhere;
//   - documents an exported identifier, or the package, in a contract
//     package (anything not inside a module's internal directory);
//   - documents a command (a main package's package doc, its manual);
//   - is a directive (//go:build, //go:generate, //line …) or a generated
//     file's header.
package comments

import (
	"go/ast"
	"strings"

	"github.com/McReaper/go-agent-kit/analyzers/internal/layout"
	"golang.org/x/tools/go/analysis"
)

const message = "M2: a comment says why, not what: start it with \"WHY:\" or delete it " +
	"(doc comments belong to the exported API of a contract package)"

// New returns the analyzer.
func New() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "comments",
		Doc:  "reports comments that are neither a WHY: nor the doc of a contract package's exported API",
		Run:  layout.Run(run),
	}
}

func run(pass *analysis.Pass) error {
	for _, f := range pass.Files {
		checkFile(pass, f)
	}
	return nil
}

func checkFile(pass *analysis.Pass, f *ast.File) {
	docs := map[*ast.CommentGroup]bool{}
	if layout.IsMain(pass.Pkg) {
		docs[f.Doc] = true
	}
	if layout.IsContract(pass.Pkg) {
		collectDocs(f, docs)
	}
	for _, cg := range f.Comments {
		if !docs[cg] && !allowed(cg) {
			pass.Reportf(cg.Pos(), message)
		}
	}
}

func allowed(cg *ast.CommentGroup) bool {
	for _, c := range cg.List {
		switch {
		case isDirective(c.Text):
		case strings.HasPrefix(bareText(c.Text), "WHY:"):
			return true
		default:
			return false
		}
	}
	return true
}

func bareText(comment string) string {
	return strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(comment, "//"), "/*"), "*/"))
}

func isDirective(text string) bool {
	for _, p := range []string{"//go:", "//line ", "//export ", "//nolint", "//lint:", "// Code generated "} {
		if strings.HasPrefix(text, p) {
			return true
		}
	}
	return false
}

func collectDocs(f *ast.File, docs map[*ast.CommentGroup]bool) {
	docs[f.Doc] = true
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			docs[d.Doc] = d.Name.IsExported()
		case *ast.GenDecl:
			genDocs(d, docs)
		}
	}
}

func genDocs(d *ast.GenDecl, docs map[*ast.CommentGroup]bool) {
	for _, s := range d.Specs {
		specDocs(d, s, docs)
	}
}

func specDocs(d *ast.GenDecl, s ast.Spec, docs map[*ast.CommentGroup]bool) {
	switch s := s.(type) {
	case *ast.TypeSpec:
		if s.Name.IsExported() {
			docs[d.Doc], docs[s.Doc], docs[s.Comment] = true, true, true
			fieldDocs(s.Type, docs)
		}
	case *ast.ValueSpec:
		if anyExported(s.Names) {
			docs[d.Doc], docs[s.Doc], docs[s.Comment] = true, true, true
		}
	}
}

func fieldDocs(t ast.Expr, docs map[*ast.CommentGroup]bool) {
	var fields *ast.FieldList
	switch t := t.(type) {
	case *ast.StructType:
		fields = t.Fields
	case *ast.InterfaceType:
		fields = t.Methods
	default:
		return
	}
	for _, fl := range fields.List {
		exported := len(fl.Names) == 0 || anyExported(fl.Names)
		docs[fl.Doc], docs[fl.Comment] = exported, exported
	}
}

func anyExported(names []*ast.Ident) bool {
	for _, n := range names {
		if n.IsExported() {
			return true
		}
	}
	return false
}
