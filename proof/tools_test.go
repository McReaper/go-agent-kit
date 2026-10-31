package proof_test

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

type tool struct {
	prefix string
	run    func(*testing.T, string) []finding
}

var standalone = []tool{
	{"actionlint/", actionlintFindings},
	{"deadcode/", deadcodeFindings},
	{"gitleaks/", gitleaksFindings},
	{"gomod/", tidyFindings},
	{"govulncheck/", govulnFindings},
	{"nilaway/", nilawayFindings},
	{"shellcheck/", shellcheckFindings},
	{"zizmor/", zizmorFindings},
}

func TestStandaloneToolsFindNothingInTheCleanTrial(t *testing.T) {
	for _, tl := range standalone {
		t.Run(strings.TrimSuffix(tl.prefix, "/"), func(t *testing.T) {
			check(t, nil, tl.prefix, tl.run(t, cleanDir))
		})
	}
}

func TestStandaloneToolsReportEveryCase(t *testing.T) {
	for _, tl := range standalone {
		t.Run(strings.TrimSuffix(tl.prefix, "/"), func(t *testing.T) {
			proveCases(t, tl.prefix, tl.run)
		})
	}
}

// WHY: a tool that exits non-zero for findings is fine; one that couldn't
// run at all fails the proof rather than passing as "nothing found".
func runIn(t *testing.T, dir string, okCodes []int, argv ...string) []byte {
	t.Helper()
	name, args := argv[0], argv[1:]
	cmd := exec.CommandContext(t.Context(), name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOFLAGS=-mod=readonly")
	out, err := cmd.Output()
	if err == nil {
		return out
	}
	if slices.Contains(okCodes, exitCode(err)) {
		return out
	}
	t.Fatalf("%s %v: %v\n%s%s", name, args, err, out, stderrOf(err))
	return nil
}

var diagLine = regexp.MustCompile(`^(.+?):(\d+)(?::\d+)?: (.*)$`)

type diagTool struct {
	prefix string
	rule   func(string) string
}

func lineDiagnostics(t *testing.T, dir string, out []byte, tl diagTool) []finding {
	t.Helper()
	var findings []finding
	for line := range strings.Lines(string(out)) {
		m := diagLine.FindStringSubmatch(strings.TrimRight(line, "\r\n"))
		if m == nil || strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t") {
			continue
		}
		file := m[1]
		if filepath.IsAbs(file) {
			file = relTo(t, dir, file)
		}
		findings = append(findings, finding{tool: tl.prefix + tl.rule(m[3]), file: filepath.ToSlash(file), line: atoi(t, m[2]), text: m[3]})
	}
	return findings
}

var bracketRule = regexp.MustCompile(`\[([A-Za-z0-9-]+)\]$`)

func lastBracket(msg string) string {
	if m := bracketRule.FindStringSubmatch(msg); m != nil {
		return m[1]
	}
	return "unknown"
}

// WHY: actionlint finds workflows through the git repository; given the
// files by name, it lints a tree that isn't one, like the proof's copies.
func actionlintFindings(t *testing.T, dir string) []finding {
	t.Helper()
	workflows := filesWithExt(t, dir, ".yml", ".yaml")
	if len(workflows) == 0 {
		t.Fatal("no workflow to check: the trial must hold at least one")
	}
	out := runIn(t, dir, []int{1}, append([]string{"actionlint", "-oneline"}, workflows...)...)
	return lineDiagnostics(t, dir, out, diagTool{"actionlint/", lastBracket})
}

func filesWithExt(t *testing.T, dir string, exts ...string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !slices.Contains(exts, filepath.Ext(path)) {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		files = append(files, filepath.ToSlash(rel))
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

var shellShebang = regexp.MustCompile(`^#!\s*(/bin/|/usr/bin/env\s+)(ba|da|k)?sh\b`)

// WHY: a git hook has no extension; like kit check, the proof knows it by
// its shebang.
func gitHooks(t *testing.T, dir string) []string {
	t.Helper()
	var hooks []string
	for _, f := range filesWithExt(t, dir, "") {
		data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(f)))
		if err != nil {
			t.Fatal(err)
		}
		if shellShebang.Match(data) {
			hooks = append(hooks, f)
		}
	}
	return hooks
}

func shellcheckFindings(t *testing.T, dir string) []finding {
	t.Helper()
	scripts := append(filesWithExt(t, dir, ".sh"), gitHooks(t, dir)...)
	if len(scripts) == 0 {
		t.Fatal("no shell script to check: the trial must hold at least one")
	}
	out := runIn(t, dir, []int{1}, append([]string{"shellcheck", "-f", "gcc"}, scripts...)...)
	return lineDiagnostics(t, dir, out, diagTool{"shellcheck/", lastBracket})
}

func deadcodeFindings(t *testing.T, dir string) []finding {
	t.Helper()
	out := runIn(t, dir, nil, "deadcode", "./cmd/...")
	return lineDiagnostics(t, dir, out, diagTool{"deadcode/", func(string) string { return "unreachable" }})
}

// WHY: nilaway prints its diagnostics to standard error, like go vet.
func nilawayFindings(t *testing.T, dir string) []finding {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), "nilaway", "-pretty-print=false", "./...")
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil && exitCode(err) != 3 {
		t.Fatalf("nilaway: %v\n%s", err, stderr.String())
	}
	return lineDiagnostics(t, dir, stderr.Bytes(), diagTool{"nilaway/", func(string) string { return "nil-panic" }})
}

func tidyFindings(t *testing.T, dir string) []finding {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), "go", "mod", "tidy", "-diff")
	cmd.Dir = dir
	out, err := cmd.Output()
	switch {
	case err == nil:
		return nil
	case exitCode(err) == 1 && len(out) > 0:
		return []finding{{tool: "gomod/tidy", file: "go.mod", line: 0, text: "go.mod or go.sum is not tidy:\n" + string(out)}}
	}
	t.Fatalf("go mod tidy -diff: %v\n%s", err, stderrOf(err))
	return nil
}

type gitleaksLeak struct {
	RuleID    string `json:"RuleID"`
	File      string `json:"File"`
	StartLine int    `json:"StartLine"`
	Desc      string `json:"Description"`
}

func gitleaksFindings(t *testing.T, dir string) []finding {
	t.Helper()
	out := runIn(t, dir, []int{1}, "gitleaks", "dir", ".", "--no-banner", "--redact", "-f", "json", "-r", "-")
	var leaks []gitleaksLeak
	if err := json.Unmarshal(out, &leaks); err != nil {
		t.Fatalf("gitleaks: %v\n%s", err, out)
	}
	findings := make([]finding, 0, len(leaks))
	for _, l := range leaks {
		findings = append(findings, finding{tool: "gitleaks/" + l.RuleID, file: filepath.ToSlash(l.File), line: l.StartLine, text: l.Desc})
	}
	return findings
}

type zizmorFinding struct {
	Ident     string `json:"ident"`
	Desc      string `json:"desc"`
	Locations []struct {
		Symbolic struct {
			Kind string `json:"kind"`
			Key  struct {
				Local struct {
					Path string `json:"verbatim_path"`
				} `json:"Local"`
			} `json:"key"`
		} `json:"symbolic"`
		Concrete struct {
			Location struct {
				Start struct {
					Row int `json:"row"`
				} `json:"start_point"`
			} `json:"location"`
		} `json:"concrete"`
	} `json:"locations"`
}

func zizmorFindings(t *testing.T, dir string) []finding {
	t.Helper()
	out := runIn(t, dir, []int{11, 12, 13, 14}, zizmorPath(t), "--offline", "--format", "json", ".")
	var zs []zizmorFinding
	if err := json.Unmarshal(out, &zs); err != nil {
		t.Fatalf("zizmor: %v\n%s", err, out)
	}
	var findings []finding
	for _, z := range zs {
		for _, l := range z.Locations {
			if l.Symbolic.Kind == "Primary" {
				file := strings.TrimPrefix(filepath.ToSlash(l.Symbolic.Key.Local.Path), "./")
				findings = append(findings, finding{tool: "zizmor/" + z.Ident, file: file, line: l.Concrete.Location.Start.Row + 1, text: z.Desc})
			}
		}
	}
	return findings
}

// WHY: pip installs zizmor outside the PATH on Windows; ZIZMOR names it.
func zizmorPath(t *testing.T) string {
	t.Helper()
	if p := os.Getenv("ZIZMOR"); p != "" {
		return p
	}
	return "zizmor"
}

type govulnMessage struct {
	Finding *struct {
		OSV          string `json:"osv"`
		FixedVersion string `json:"fixed_version"`
		Trace        []struct {
			Function string `json:"function"`
			Position *struct {
				Filename string `json:"filename"`
				Line     int    `json:"line"`
			} `json:"position"`
		} `json:"trace"`
	} `json:"finding"`
}

// WHY: govulncheck reports a vulnerability in a dependency only when the
// code calls it; the last frame of a called trace is where the code does.
func govulnFindings(t *testing.T, dir string) []finding {
	t.Helper()
	out := runIn(t, dir, nil, "govulncheck", "-format", "json", "./...")
	dec := json.NewDecoder(bytes.NewReader(out))
	var findings []finding
	for dec.More() {
		var m govulnMessage
		if err := dec.Decode(&m); err != nil {
			t.Fatalf("govulncheck: %v", err)
		}
		if f, ok := calledVuln(m); ok && !containsFinding(findings, f) {
			findings = append(findings, f)
		}
	}
	return findings
}

func calledVuln(m govulnMessage) (finding, bool) {
	if m.Finding == nil || len(m.Finding.Trace) == 0 || m.Finding.Trace[0].Function == "" {
		return finding{}, false
	}
	last := m.Finding.Trace[len(m.Finding.Trace)-1]
	if last.Position == nil {
		return finding{}, false
	}
	return finding{
		tool: "govulncheck/" + m.Finding.OSV, file: filepath.ToSlash(last.Position.Filename), line: last.Position.Line,
		text: "called; fixed in " + m.Finding.FixedVersion,
	}, true
}

func containsFinding(fs []finding, f finding) bool {
	return slices.Contains(fs, f)
}

func atoi(t *testing.T, s string) int {
	t.Helper()
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			t.Fatalf("%q is not a line number", s)
		}
		n = n*10 + int(r-'0')
	}
	return n
}
