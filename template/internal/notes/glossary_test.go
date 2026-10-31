package notes_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// WHY: the glossary is the module's words for people and agents; a domain
// type it doesn't define is a word each reader guesses at.
func TestTheGlossaryDefinesEveryDomainType(t *testing.T) {
	glossary, err := os.ReadFile("GLOSSARY.md")
	if err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join("internal", "domain", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, file, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range exportedTypes(f) {
			if !strings.Contains(string(glossary), "- **"+name+"**: ") {
				t.Errorf("GLOSSARY.md doesn't define %s (%s): add a line \"- **%s**: what it means here\"", name, file, name)
			}
		}
	}
}

func exportedTypes(f *ast.File) []string {
	var out []string
	for _, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			if ts, ok := spec.(*ast.TypeSpec); ok && ts.Name.IsExported() {
				out = append(out, ts.Name.Name)
			}
		}
	}
	return out
}
