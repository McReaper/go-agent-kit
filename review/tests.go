package review

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/McReaper/go-agent-kit/guard"
	"github.com/McReaper/go-agent-kit/internal/platform/git"
)

// Tests are the test scenarios a change adds, removes, or changes the body
// of, each read as a sentence.
type Tests struct {
	Added, Removed, Changed []Item
}

func (t Tests) empty() bool {
	return len(t.Added)+len(t.Removed)+len(t.Changed) == 0
}

func scenarios(ctx context.Context, dir string, d guard.Diff) (t Tests, err error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return Tests{}, fmt.Errorf("open %s: %w", dir, err)
	}
	defer func() { err = errors.Join(err, root.Close()) }()
	src := sources{dir: dir, root: root, base: d.Base}
	for _, f := range d.Files {
		if !strings.HasSuffix(f.Path, "_test.go") {
			continue
		}
		before, err := src.before(ctx, f)
		if err != nil {
			return Tests{}, err
		}
		after, err := src.after(f)
		if err != nil {
			return Tests{}, err
		}
		t.compare(f.Path, before, after)
	}
	return t, nil
}

type sources struct {
	dir  string
	root *os.Root
	base string
}

func (s sources) before(ctx context.Context, f guard.File) (map[string]string, error) {
	if f.Status == 'A' {
		return map[string]string{}, nil
	}
	old := f.Path
	if f.OldPath != "" {
		old = f.OldPath
	}
	src, err := git.Run(ctx, s.dir, "show", s.base+":"+old)
	if err != nil {
		return nil, err
	}
	return testsIn(old, []byte(src))
}

func (s sources) after(f guard.File) (map[string]string, error) {
	if f.Status == 'D' {
		return map[string]string{}, nil
	}
	src, err := s.root.ReadFile(f.Path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", f.Path, err)
	}
	return testsIn(f.Path, src)
}

func (t *Tests) compare(file string, before, after map[string]string) {
	for _, name := range slices.Sorted(maps.Keys(after)) {
		old, ok := before[name]
		switch {
		case !ok:
			t.Added = append(t.Added, Item{File: file, Text: Sentence(name)})
		case old != after[name] && !strings.Contains(name, "/"):
			t.Changed = append(t.Changed, Item{File: file, Text: Sentence(name)})
		}
	}
	for _, name := range slices.Sorted(maps.Keys(before)) {
		if _, ok := after[name]; !ok {
			t.Removed = append(t.Removed, Item{File: file, Text: Sentence(name)})
		}
	}
}

// WHY: a subtest named by a variable (a table's rows) is known only when
// it runs; its test shows as changed instead.
func testsIn(file string, src []byte) (map[string]string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, src, parser.SkipObjectResolution)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", file, err)
	}
	out := map[string]string{}
	text := func(n ast.Node) string {
		return string(src[fset.Position(n.Pos()).Offset:fset.Position(n.End()).Offset])
	}
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || !isTestName(fn.Name.Name) {
			continue
		}
		out[fn.Name.Name] = text(fn)
		for name, n := range subtests(fn) {
			out[fn.Name.Name+"/"+name] = text(n)
		}
	}
	return out, nil
}

func subtests(fn *ast.FuncDecl) map[string]ast.Node {
	out := map[string]ast.Node{}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || len(call.Args) != 2 {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		lit, isLit := call.Args[0].(*ast.BasicLit)
		if !ok || sel.Sel.Name != "Run" || !isLit || lit.Kind != token.STRING {
			return true
		}
		if name, err := strconv.Unquote(lit.Value); err == nil {
			out[name] = call
		}
		return true
	})
	return out
}

// WHY: go test's own rule: the prefix, then nothing or a character that
// isn't a lower-case letter (TestX runs, Testx doesn't).
func isTestName(name string) bool {
	for _, prefix := range []string{"Test", "Fuzz", "Example"} {
		rest, ok := strings.CutPrefix(name, prefix)
		if !ok {
			continue
		}
		if rest == "" || !unicode.IsLower([]rune(rest)[0]) {
			return true
		}
	}
	return false
}

// Sentence reads a test's name as a sentence: TestAddRefusesAnEmptyTitle
// is "Add refuses an empty title", a subtest follows a colon, fuzz tests
// and examples are labelled.
func Sentence(name string) string {
	fn, sub, hasSub := strings.Cut(name, "/")
	label := ""
	switch {
	case strings.HasPrefix(fn, "Test"):
		fn = fn[len("Test"):]
	case strings.HasPrefix(fn, "Fuzz"):
		fn, label = fn[len("Fuzz"):], "Fuzz: "
	case strings.HasPrefix(fn, "Example"):
		fn, label = fn[len("Example"):], "Example: "
	}
	s := label + capitalize(strings.Join(words(fn), " "))
	if hasSub {
		s += ": " + strings.ReplaceAll(sub, "_", " ")
	}
	return s
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func words(id string) []string {
	var out []string
	for part := range strings.SplitSeq(id, "_") {
		for _, w := range camel([]rune(part)) {
			out = append(out, inSentence(w))
		}
	}
	return out
}

// WHY: an acronym (JSON, ID) stays upper case; every other word reads lower
// case in a sentence.
func inSentence(w string) string {
	if len(w) > 1 && strings.ToUpper(w) == w {
		return w
	}
	return strings.ToLower(w)
}

func camel(r []rune) []string {
	var out []string
	start := 0
	for i := 1; i < len(r); i++ {
		if startsWord(r, i) {
			out = append(out, string(r[start:i]))
			start = i
		}
	}
	if start < len(r) {
		out = append(out, string(r[start:]))
	}
	return out
}

// WHY: an upper-case letter starts a word after a lower-case letter or a
// digit, and ends an acronym when a lower-case letter follows it (JSONFile
// is JSON, File).
func startsWord(r []rune, i int) bool {
	if !unicode.IsUpper(r[i]) {
		return false
	}
	prev := r[i-1]
	if unicode.IsLower(prev) || unicode.IsDigit(prev) {
		return true
	}
	return unicode.IsUpper(prev) && i+1 < len(r) && unicode.IsLower(r[i+1])
}
