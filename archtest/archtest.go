// Package archtest checks a Go module against the kit's layout (criteria
// E1, E2 and E7):
//
//	cmd/<app>                          wiring only
//	internal/<m>                       the module's contract
//	internal/<m>/internal/domain       pure: the standard library and pkg/ only
//	internal/<m>/internal/infra        I/O adapters
//	internal/<m>/internal/<feature>    a vertical slice
//	internal/platform/<p>              the shared kernel, which knows no module
//	pkg/<lib>                          libraries
//
// The compiler already keeps other modules out of a module's internal
// directory; archtest adds what it can't see. An app calls Test from one
// test of its own.
package archtest

import (
	"fmt"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"testing"

	"golang.org/x/tools/go/packages"
)

// Violation is one import, or one package, the layout forbids.
type Violation struct {
	Rule    string
	File    string
	Line    int
	Message string
}

// Test fails t with each violation found in the module at dir.
func Test(tb testing.TB, dir string) {
	tb.Helper()
	vs, err := Check(dir)
	if err != nil {
		tb.Fatal(err)
	}
	for _, v := range vs {
		tb.Errorf("%s:%d: %s: %s", v.File, v.Line, v.Rule, v.Message)
	}
}

// Check returns the violations in the module at dir, its tests aside.
func Check(dir string) ([]Violation, error) {
	cfg := &packages.Config{Mode: packages.NeedName | packages.NeedFiles | packages.NeedModule, Dir: dir}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}
	var out []Violation
	for _, p := range pkgs {
		if len(p.Errors) > 0 {
			return nil, fmt.Errorf("load %s: %w", p.PkgPath, p.Errors[0])
		}
		vs, err := checkPackage(p)
		if err != nil {
			return nil, err
		}
		out = append(out, vs...)
	}
	return out, nil
}

func checkPackage(p *packages.Package) ([]Violation, error) {
	if p.Module == nil {
		return nil, fmt.Errorf("%s is not in a module", p.PkgPath)
	}
	from := classify(p.Module.Path, p.PkgPath)
	var out []Violation
	if from.kind == unplaced && len(p.GoFiles) > 0 {
		out = append(out, Violation{"layout", p.GoFiles[0], 1, fmt.Sprintf(
			"E7: package %s is outside the layout: a module's packages live in internal/<m>/internal/<name>, "+
				"shared code in internal/platform/<name>, libraries in pkg/<name>", p.PkgPath)})
	}
	for _, file := range p.GoFiles {
		vs, err := checkImports(p.Module.Path, from, file)
		if err != nil {
			return nil, err
		}
		out = append(out, vs...)
	}
	return out, nil
}

func checkImports(module string, from place, file string) ([]Violation, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, parser.ImportsOnly)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", filepath.Base(file), err)
	}
	var out []Violation
	for _, spec := range f.Imports {
		path, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return nil, fmt.Errorf("%s: import %s: %w", file, spec.Path.Value, err)
		}
		if rule, msg, bad := forbidden(module, from, path); bad {
			out = append(out, Violation{rule, file, fset.Position(spec.Pos()).Line, msg})
		}
	}
	return out, nil
}
