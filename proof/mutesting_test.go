package proof_test

import (
	"regexp"
	"strings"
	"testing"
)

var (
	hunkHeader  = regexp.MustCompile(`^@@ -(\d+)`)
	mutantFile  = regexp.MustCompile(`([^/\\]+\.go)\.\d+"`)
	lineOfDiffs = regexp.MustCompile(`^[-+]\t`)
)

// WHY: go-mutesting writes each mutant over the original file while the
// tests run; it only ever runs on an overlay copy, never on template.
func mutestingFindings(t *testing.T, dir string) []finding {
	t.Helper()
	var findings []finding
	for _, pkg := range casePackages(t, dir) {
		out := runIn(t, dir, []int{1}, "go-mutesting", "./"+pkg)
		findings = append(findings, livedMutants(pkg, string(out))...)
	}
	return findings
}

// WHY: mutation testing is slow, so the proof mutates only the packages a
// case adds, named *case.
func casePackages(t *testing.T, dir string) []string {
	t.Helper()
	out := runIn(t, dir, nil, "go", "list", "-f", "{{.Dir}}", "./...")
	var pkgs []string
	for line := range strings.Lines(string(out)) {
		d := strings.TrimSpace(line)
		if strings.HasSuffix(d, "case") {
			pkgs = append(pkgs, relTo(t, dir, d))
		}
	}
	return pkgs
}

type mutantReader struct {
	pkg      string
	hunk     int
	offset   int
	changed  int
	mutation string
}

func livedMutants(pkg, out string) []finding {
	r := mutantReader{pkg: pkg}
	var findings []finding
	for line := range strings.Lines(out) {
		if f, ok := r.line(strings.TrimRight(line, "\r\n")); ok {
			findings = append(findings, f)
		}
	}
	return findings
}

func (r *mutantReader) line(line string) (finding, bool) {
	switch {
	case hunkHeader.MatchString(line):
		r.hunk, r.offset, r.changed, r.mutation = atoiUnsafe(hunkHeader.FindStringSubmatch(line)[1]), 0, 0, ""
	case strings.HasPrefix(line, "FAIL "):
		file := mutantFile.FindStringSubmatch(line)
		if file == nil {
			return finding{}, false
		}
		return finding{tool: "mutesting/lived", file: r.pkg + "/" + file[1], line: r.changed, text: r.mutation}, true
	case lineOfDiffs.MatchString(line) && r.changed == 0 && line[0] == '-':
		r.changed = r.hunk + r.offset
	case lineOfDiffs.MatchString(line) && line[0] == '+':
		r.mutation = strings.TrimSpace(line[1:])
	case strings.HasPrefix(line, " ") || line == "":
		r.offset++
	}
	return finding{}, false
}

func atoiUnsafe(s string) int {
	n := 0
	for _, r := range s {
		n = n*10 + int(r-'0')
	}
	return n
}

func TestMutationReportsEveryLivedMutant(t *testing.T) {
	proveCases(t, "mutesting/", mutestingFindings)
}
