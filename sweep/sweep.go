// Package sweep checks what only the whole module shows: code only tests
// reach (criterion M5, where deadcode never reports a method), exported
// names their own package alone uses (M5), log messages written in two
// places (M7, the unique EventId of .NET), and function bodies written
// twice (M4).
package sweep

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Finding is one rule the module breaks.
type Finding struct {
	Rule    string
	File    string
	Line    int
	Message string
}

// Check loads the module at dir, its tests included, and returns what it
// breaks.
func Check(dir string) ([]Finding, error) {
	cfg := &packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedModule,
		Dir:   dir,
		Tests: true,
	}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}
	if n := packages.PrintErrors(pkgs); n > 0 {
		return nil, fmt.Errorf("load packages: %d errors", n)
	}
	m := newModel(pkgs)
	out := m.onlyTests()
	out = append(out, m.overExported()...)
	out = append(out, m.duplicateMessages()...)
	return append(out, m.copiedFunctions()...), nil
}

type use struct {
	pkg  string
	test bool
}

type model struct {
	fset      *token.FileSet
	decls     map[types.Object]*ast.Ident
	uses      map[types.Object][]use
	ifaceSeen map[string]bool
	messages  map[string][]token.Position
	bodies    map[string][]site
}

func newModel(pkgs []*packages.Package) *model {
	m := &model{decls: map[types.Object]*ast.Ident{}, uses: map[types.Object][]use{}, ifaceSeen: map[string]bool{}, messages: map[string][]token.Position{}, bodies: map[string][]site{}}
	packages.Visit(pkgs, nil, func(p *packages.Package) {
		if p.Module == nil || !p.Module.Main || p.TypesInfo == nil {
			return
		}
		m.fset = p.Fset
		m.read(p)
	})
	return m
}

func (m *model) read(p *packages.Package) {
	for _, f := range p.Syntax {
		test := strings.HasSuffix(p.Fset.File(f.Pos()).Name(), "_test.go")
		if !test && !isTestVariant(p) {
			m.declare(p, f)
			m.collectMessages(p, f)
			m.collectBodies(p, f)
		}
		m.collectUses(p, f, test)
	}
	if !isTestVariant(p) {
		m.collectInterfaces(p)
	}
}

// WHY: with Tests set, go/packages loads a package twice, plain and with
// its in-package tests; its declarations are counted from the plain one.
func isTestVariant(p *packages.Package) bool {
	return strings.Contains(p.ID, " [") || strings.HasSuffix(p.PkgPath, "_test") || strings.HasSuffix(p.PkgPath, ".test")
}

func (m *model) declare(p *packages.Package, f *ast.File) {
	for id, obj := range p.TypesInfo.Defs {
		if obj == nil || id.Pos() < f.Pos() || id.Pos() > f.End() {
			continue
		}
		if isTopLevel(obj) || isMethod(obj) {
			m.decls[obj] = id
		}
	}
}

func (m *model) collectUses(p *packages.Package, f *ast.File, test bool) {
	for id, obj := range p.TypesInfo.Uses {
		if id.Pos() >= f.Pos() && id.Pos() <= f.End() && obj.Pkg() != nil {
			m.uses[origin(obj)] = append(m.uses[origin(obj)], use{pkg: p.PkgPath, test: test || isTestVariant(p) && strings.HasSuffix(p.PkgPath, "_test")})
		}
	}
}

func (m *model) collectInterfaces(p *packages.Package) {
	for _, sel := range p.TypesInfo.Selections {
		if types.IsInterface(sel.Recv()) {
			m.ifaceSeen[sel.Obj().Name()] = true
		}
	}
	for _, tv := range p.TypesInfo.Types {
		markInterfaceMethods(tv.Type, m.ifaceSeen)
	}
}

func markInterfaceMethods(t types.Type, seen map[string]bool) {
	iface, ok := t.Underlying().(*types.Interface)
	if !ok {
		return
	}
	for f := range iface.Methods() {
		seen[f.Name()] = true
	}
}

// WHY: generic functions and methods of generic types are used through
// their instances; the declared object is their origin.
func origin(obj types.Object) types.Object {
	switch o := obj.(type) {
	case *types.Func:
		return o.Origin()
	case *types.Var:
		return o.Origin()
	}
	return obj
}

func isTopLevel(obj types.Object) bool {
	return obj.Parent() != nil && obj.Parent() == obj.Pkg().Scope()
}

func isMethod(obj types.Object) bool {
	fn, ok := obj.(*types.Func)
	return ok && fn.Signature().Recv() != nil
}
