package check

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/McReaper/rumblestrip-go/internal/platform/run"
)

// WHY: each tool says "found something" with its own exit code (1 for
// golangci-lint, 3 for nilaway and govulncheck); any other code means it
// could not run, which fails the gate rather than passing it silently.
type ran struct {
	out run.Output
	err error
}

func (r ran) result(findingCode int, text []byte) Result {
	switch {
	case r.err != nil:
		return failed(r.err)
	case r.out.Code == 0:
		return Result{}
	case r.out.Code == findingCode:
		return found(nonEmpty(text))
	}
	return failed(fmt.Errorf("exit %d: %s", r.out.Code, strings.TrimSpace(string(r.out.Stderr))))
}

func deadcodeGate(ctx context.Context, o Options) Result {
	if fi, err := os.Stat(filepath.Join(o.Dir, "cmd")); err != nil || !fi.IsDir() {
		return Result{}
	}
	out, err := run.Command(ctx, o.Dir, "deadcode", "./cmd/...")
	if err != nil {
		return failed(err)
	}
	if out.Code != 0 {
		return failed(fmt.Errorf("exit %d: %s", out.Code, strings.TrimSpace(string(out.Stderr))))
	}
	return found(nonEmpty(out.Stdout))
}

func nilawayGate(ctx context.Context, o Options) Result {
	out, err := run.Command(ctx, o.Dir, "nilaway", "-pretty-print=false", "./...")
	return ran{out, err}.result(3, out.Stderr)
}

func vulnGate(ctx context.Context, o Options) Result {
	out, err := run.Command(ctx, o.Dir, "govulncheck", "./...")
	r := ran{out, err}.result(3, out.Stdout)
	r.Findings = keep(r.Findings, "Vulnerability #")
	return r
}

func testGate(ctx context.Context, o Options) Result {
	argv := []string{"go", "test", "-json", "-count=1"}
	if o.Race {
		argv = append(argv, "-race")
	}
	out, err := run.Command(ctx, o.Dir, append(argv, "./...")...)
	switch {
	case err != nil:
		return failed(err)
	case out.Code == 0:
		return Result{}
	case out.Code != 1:
		return failed(fmt.Errorf("exit %d: %s", out.Code, strings.TrimSpace(string(out.Stderr))))
	}
	if fs := testFindings(o.Dir, out.Stdout); len(fs) > 0 {
		return Result{Findings: fs}
	}
	return failed(fmt.Errorf("go test failed and named no test: %s", strings.TrimSpace(string(out.Stderr))))
}

func keep(lines []string, prefixes ...string) []string {
	var out []string
	for _, l := range lines {
		if slices.ContainsFunc(prefixes, func(p string) bool { return strings.HasPrefix(l, p) }) {
			out = append(out, l)
		}
	}
	return out
}
