package check

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/McReaper/go-agent-kit/internal/platform/run"
)

// WHY: each tool says "found something" with its own exit code (1 for
// golangci-lint, 3 for nilaway and govulncheck); any other code means it
// could not run, which fails the gate rather than passing it silently.
func fromLines(out run.Output, err error, findingCode int, text []byte) Result {
	switch {
	case err != nil:
		return failed(err)
	case out.Code == 0:
		return Result{}
	case out.Code == findingCode:
		return found(nonEmpty(text))
	}
	return failed(fmt.Errorf("exit %d: %s", out.Code, strings.TrimSpace(string(out.Stderr))))
}

func deadcodeGate(ctx context.Context, o Options) Result {
	if fi, err := os.Stat(filepath.Join(o.Dir, "cmd")); err != nil || !fi.IsDir() {
		return Result{}
	}
	out, err := run.Command(ctx, o.Dir, nil, "deadcode", "./cmd/...")
	if err != nil {
		return failed(err)
	}
	if out.Code != 0 {
		return failed(fmt.Errorf("exit %d: %s", out.Code, strings.TrimSpace(string(out.Stderr))))
	}
	return found(nonEmpty(out.Stdout))
}

func nilawayGate(ctx context.Context, o Options) Result {
	out, err := run.Command(ctx, o.Dir, nil, "nilaway", "-pretty-print=false", "./...")
	return fromLines(out, err, 3, out.Stderr)
}

func vulnGate(ctx context.Context, o Options) Result {
	out, err := run.Command(ctx, o.Dir, nil, "govulncheck", "./...")
	r := fromLines(out, err, 3, out.Stdout)
	r.Findings = keep(r.Findings, "Vulnerability #")
	return r
}

func testGate(ctx context.Context, o Options) Result {
	argv := []string{"go", "test", "-count=1"}
	if o.Race {
		argv = append(argv, "-race")
	}
	out, err := run.Command(ctx, o.Dir, nil, append(argv, "./...")...)
	r := fromLines(out, err, 1, append(out.Stdout, out.Stderr...))
	r.Findings = keep(r.Findings, "--- FAIL", "FAIL", "panic:", "#")
	return r
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
