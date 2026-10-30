package proof

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

const (
	cleanDir = "../trial/clean"
	casesDir = "../trial/cases"
)

type want struct {
	caseName string
	tool     string
	file     string
	line     int
	re       *regexp.Regexp
}

func (w want) matches(f finding) bool {
	return w.tool == f.tool && w.file == f.file && w.line == f.line && w.re.MatchString(f.text)
}

func (w want) String() string {
	return fmt.Sprintf("%s %s:%d %s", w.tool, w.file, w.line, w.re)
}

type finding struct {
	tool string
	file string
	line int
	text string
}

func (f finding) String() string {
	return fmt.Sprintf("%s %s:%d %s", f.tool, f.file, f.line, f.text)
}

type trialCase struct {
	name     string
	dir      string
	wants    []want
	isolated bool
	deletes  []string
}

func (c trialCase) uses(toolPrefix string) bool {
	return slices.ContainsFunc(c.wants, func(w want) bool { return strings.HasPrefix(w.tool, toolPrefix) })
}

func loadCases(t *testing.T) []trialCase {
	t.Helper()
	entries, err := os.ReadDir(casesDir)
	if err != nil {
		t.Fatal(err)
	}
	cases := make([]trialCase, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			cases = append(cases, loadCase(t, e.Name()))
		}
	}
	return cases
}

func loadCase(t *testing.T, name string) trialCase {
	t.Helper()
	c := trialCase{name: name, dir: filepath.Join(casesDir, name)}
	f, err := os.Open(filepath.Join(c.dir, "want.txt"))
	if err != nil {
		t.Fatalf("case %s: %v", name, err)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		switch {
		case line == "" || strings.HasPrefix(line, "#"):
		case line == "!isolated":
			c.isolated = true
		case strings.HasPrefix(line, "!delete "):
			c.deletes = append(c.deletes, strings.TrimPrefix(line, "!delete "))
		default:
			c.wants = append(c.wants, parseWant(t, name, n, line))
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatalf("case %s: %v", name, err)
	}
	return c
}

func parseWant(t *testing.T, caseName string, n int, line string) want {
	t.Helper()
	fields := strings.SplitN(line, " ", 3)
	if len(fields) != 3 {
		t.Fatalf("case %s want.txt:%d: want \"<tool> <file>:<line> <regexp>\", got %q", caseName, n, line)
	}
	file, lineStr, ok := strings.Cut(fields[1], ":")
	lineNo, err := strconv.Atoi(lineStr)
	if !ok || err != nil {
		t.Fatalf("case %s want.txt:%d: %q is not <file>:<line>", caseName, n, fields[1])
	}
	re, err := regexp.Compile(fields[2])
	if err != nil {
		t.Fatalf("case %s want.txt:%d: %v", caseName, n, err)
	}
	return want{caseName: caseName, tool: fields[0], file: file, line: lineNo, re: re}
}

// overlay copies trial/clean into a new directory and applies the cases to
// it: each case's files (except want.txt) are added or replace clean ones.
func overlay(t *testing.T, cases ...trialCase) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.CopyFS(dir, os.DirFS(cleanDir)); err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		for _, d := range c.deletes {
			if err := os.Remove(filepath.Join(dir, filepath.FromSlash(d))); err != nil {
				t.Fatalf("case %s: %v", c.name, err)
			}
		}
		copyCase(t, c, dir)
	}
	return dir
}

func copyCase(t *testing.T, c trialCase, dst string) {
	t.Helper()
	err := filepath.WalkDir(c.dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || path == filepath.Join(c.dir, "want.txt") {
			return err
		}
		rel, err := filepath.Rel(c.dir, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o600)
	})
	if err != nil {
		t.Fatalf("case %s: %v", c.name, err)
	}
}

// check fails unless every finding matches exactly one want of the cases,
// and every want is matched.
func check(t *testing.T, cases []trialCase, toolPrefix string, findings []finding) {
	t.Helper()
	var wants []want
	for _, c := range cases {
		for _, w := range c.wants {
			if strings.HasPrefix(w.tool, toolPrefix) {
				wants = append(wants, w)
			}
		}
	}
	used := make([]bool, len(wants))
	for _, f := range findings {
		i := slices.IndexFunc(wants, func(w want) bool { return w.matches(f) })
		for i >= 0 && used[i] {
			next := slices.IndexFunc(wants[i+1:], func(w want) bool { return w.matches(f) })
			if next < 0 {
				i = -1
			} else {
				i += 1 + next
			}
		}
		if i < 0 {
			t.Errorf("unexpected finding: %s", f)
			continue
		}
		used[i] = true
	}
	for i, w := range wants {
		if !used[i] {
			t.Errorf("case %s: missing finding: %s", w.caseName, w)
		}
	}
}
