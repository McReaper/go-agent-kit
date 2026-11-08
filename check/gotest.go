package check

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/McReaper/rumblestrip-go/internal/platform/gotest"
)

var (
	testMessage = regexp.MustCompile(`^\s+(\w[\w.-]*\.go):(\d+): (.*)$`)
	buildError  = regexp.MustCompile(`^([^\s:#][^\s:]*\.go):(\d+):(?:\d+:)? (.*)$`)
	modulePath  = regexp.MustCompile(`(?m)^module\s+(\S+)`)
)

// WHY: go test's own lines say which package and test failed, not why: a
// finding is each failing test's message, at the line of the check that
// failed, so the owner's summary and the annotation say what broke.
func testFindings(dir string, stream []byte) []string {
	r := testRun{dir: dir, module: module(dir), output: map[string][]string{}}
	for _, e := range gotest.Events(stream) {
		r.read(e)
	}
	return r.findings()
}

type testRun struct {
	dir, module string
	output      map[string][]string
	failed      []gotest.Event
	built       []string
}

func (r *testRun) read(e gotest.Event) {
	key := e.Package + " " + e.Test
	switch e.Action {
	case "output":
		r.output[key] = append(r.output[key], strings.TrimRight(e.Output, "\n"))
	case "build-output":
		if m := buildError.FindStringSubmatch(strings.TrimRight(e.Output, "\n")); m != nil {
			r.built = append(r.built, fmt.Sprintf("%s:%s: build: %s", m[1], m[2], m[3]))
		}
	case "fail":
		r.failed = append(r.failed, e)
	}
}

func (r *testRun) findings() []string {
	out := slices.Clone(r.built)
	for _, e := range r.failed {
		lines := r.messages(e)
		if len(lines) == 0 && !r.hasFailedChild(e) {
			lines = []string{r.unexplained(e)}
		}
		out = append(out, lines...)
	}
	return out
}

func (r *testRun) messages(e gotest.Event) []string {
	if e.Test == "" {
		return nil
	}
	var out []string
	for _, l := range r.output[e.Package+" "+e.Test] {
		m := testMessage.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		file := path.Join(r.packageDir(e.Package), m[1])
		if _, err := os.Stat(filepath.Join(r.dir, filepath.FromSlash(file))); err == nil {
			out = append(out, fmt.Sprintf("%s:%s: %s: %s", file, m[2], e.Test, m[3]))
		} else {
			out = append(out, fmt.Sprintf("%s: %s: %s:%s: %s", r.packageDir(e.Package), e.Test, m[1], m[2], m[3]))
		}
	}
	return out
}

func (r *testRun) hasFailedChild(e gotest.Event) bool {
	return slices.ContainsFunc(r.failed, func(f gotest.Event) bool {
		return f.Package == e.Package && (e.Test == "" && f.Test != "" || e.Test != "" && strings.HasPrefix(f.Test, e.Test+"/"))
	})
}

// WHY: a panic, a timeout or a TestMain that exits says nothing at a line;
// its last words are the reason.
func (r *testRun) unexplained(e gotest.Event) string {
	what := "the package's tests"
	if e.Test != "" {
		what = e.Test
	}
	last := ""
	for _, l := range r.output[e.Package+" "+e.Test] {
		if t := strings.TrimSpace(l); t != "" && !strings.HasPrefix(t, "---") && !strings.HasPrefix(t, "FAIL") && !strings.HasPrefix(t, "=== ") {
			last = t
		}
	}
	if last == "" {
		return fmt.Sprintf("%s: %s failed and said nothing", r.packageDir(e.Package), what)
	}
	return fmt.Sprintf("%s: %s failed: %s", r.packageDir(e.Package), what, last)
}

func (r *testRun) packageDir(pkg string) string {
	if rest, ok := strings.CutPrefix(pkg, r.module); ok {
		if rest = strings.TrimPrefix(rest, "/"); rest != "" {
			return rest
		}
		return "."
	}
	return pkg
}

func module(dir string) string {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return ""
	}
	data, err := root.ReadFile("go.mod")
	if err := errors.Join(err, root.Close()); err != nil {
		return ""
	}
	if m := modulePath.FindSubmatch(data); m != nil {
		return string(m[1])
	}
	return ""
}
