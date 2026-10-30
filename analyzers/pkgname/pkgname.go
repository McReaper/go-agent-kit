// Package pkgname reports package names that say nothing about their
// content (criterion E7: a utils package attracts everything no one placed)
// and exported names that repeat their package (criterion M2: callers read
// notes.NotesModule).
package pkgname

import (
	"go/ast"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/McReaper/go-agent-kit/analyzers/internal/layout"
	"golang.org/x/tools/go/analysis"
)

// New returns the analyzer.
func New() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "pkgname",
		Doc:  "reports vague package names (utils, common…) and exported names that repeat their package's",
		Run:  layout.Run(run),
	}
}

func isVague(name string) bool {
	switch name {
	case "util", "utils", "common", "helper", "helpers", "misc", "shared", "base", "lib", "types", "models", "stuff":
		return true
	}
	return false
}

func run(pass *analysis.Pass) error {
	name := pass.Pkg.Name()
	if len(pass.Files) > 0 && isVague(name) {
		pass.Reportf(pass.Files[0].Name.Pos(), "E7: package %s says nothing about what it holds: name it for what it provides, "+
			"or move each piece next to the code that uses it", name)
	}
	for _, f := range pass.Files {
		for _, id := range topLevelNames(f) {
			reportStutter(pass, name, id)
		}
	}
	return nil
}

func reportStutter(pass *analysis.Pass, pkg string, id *ast.Ident) {
	if stutters(pkg, id.Name) {
		pass.Reportf(id.Pos(), "M2: %s.%s repeats the package name: call it %s.%s", pkg, id.Name, pkg, id.Name[len(pkg):])
	}
}

func topLevelNames(f *ast.File) []*ast.Ident {
	var ids []*ast.Ident
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			ids = append(ids, funcName(d)...)
		case *ast.GenDecl:
			ids = append(ids, specNames(d.Specs)...)
		}
	}
	return ids
}

func funcName(d *ast.FuncDecl) []*ast.Ident {
	if d.Recv != nil {
		return nil
	}
	return []*ast.Ident{d.Name}
}

func specNames(specs []ast.Spec) []*ast.Ident {
	var ids []*ast.Ident
	for _, s := range specs {
		switch s := s.(type) {
		case *ast.TypeSpec:
			ids = append(ids, s.Name)
		case *ast.ValueSpec:
			ids = append(ids, s.Names...)
		}
	}
	return ids
}

func stutters(pkg, name string) bool {
	if !ast.IsExported(name) || len(name) <= len(pkg) || !strings.EqualFold(name[:len(pkg)], pkg) {
		return false
	}
	next, _ := utf8.DecodeRuneInString(name[len(pkg):])
	return unicode.IsUpper(next)
}
