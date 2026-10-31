package sweep

import (
	"cmp"
	"errors"
	"fmt"
	"go/ast"
	"go/scanner"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/tools/go/packages"
)

// WHY: one- and two-line bodies (a delegation, a getter) repeat by nature;
// a rule takes three statements or more.
const minCopiedStatements = 3

type site struct {
	name string
	pos  token.Position
}

func (m *model) collectBodies(p *packages.Package, f *ast.File) {
	src, err := source(p.Fset.File(f.Pos()).Name())
	if err != nil {
		return
	}
	for _, decl := range f.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Body == nil || len(fd.Body.List) < minCopiedStatements {
			continue
		}
		body := tokens(src[p.Fset.Position(fd.Body.Pos()).Offset:p.Fset.Position(fd.Body.End()).Offset])
		m.bodies[body] = append(m.bodies[body], site{fd.Name.Name, p.Fset.Position(fd.Name.Pos())})
	}
}

func source(file string) (data []byte, err error) {
	root, err := os.OpenRoot(filepath.Dir(file))
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", filepath.Dir(file), err)
	}
	defer func() { err = errors.Join(err, root.Close()) }()
	return root.ReadFile(filepath.Base(file))
}

// WHY: the tokens, without comments or spacing, are the body as the
// compiler reads it: a copy reformatted or commented differently is still
// the same rule.
func tokens(src []byte) string {
	var s scanner.Scanner
	fset := token.NewFileSet()
	s.Init(fset.AddFile("", -1, len(src)), src, nil, 0)
	var b strings.Builder
	for {
		_, tok, lit := s.Scan()
		if tok == token.EOF {
			return b.String()
		}
		b.WriteString(tok.String() + lit + " ")
	}
}

// WHY: dupl compares runs of 100 tokens or more, so a short rule copied
// whole into a second slice passes it (a crash test copied a ten-line
// limit check from list to search); the same body twice is the same rule
// in two homes, and the next change reaches one of them.
func (m *model) copiedFunctions() []Finding {
	var out []Finding
	for _, sites := range m.bodies {
		slices.SortFunc(sites, func(a, b site) int {
			return cmp.Or(cmp.Compare(a.pos.Filename, b.pos.Filename), cmp.Compare(a.pos.Line, b.pos.Line))
		})
		first := sites[0]
		for _, s := range sites[1:] {
			out = append(out, Finding{"copied-function", s.pos.Filename, s.pos.Line, fmt.Sprintf(
				"M4: %s repeats the body of %s (%s:%d): give the rule one home, the domain or the platform, and call it from both",
				s.name, first.name, shortPath(first.pos.Filename), first.pos.Line)})
		}
	}
	return sorted(out)
}

func shortPath(file string) string {
	return filepath.ToSlash(filepath.Join(filepath.Base(filepath.Dir(file)), filepath.Base(file)))
}
