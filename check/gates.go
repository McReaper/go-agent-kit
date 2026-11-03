package check

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/McReaper/go-agent-kit/analyzers/all"
	"github.com/McReaper/go-agent-kit/archtest"
	"github.com/McReaper/go-agent-kit/guard"
	"github.com/McReaper/go-agent-kit/internal/platform/run"
	"github.com/McReaper/go-agent-kit/lint"
	"github.com/McReaper/go-agent-kit/spec"
	"github.com/McReaper/go-agent-kit/sweep"
	"golang.org/x/tools/go/analysis/checker"
	"golang.org/x/tools/go/packages"
)

// WHY: the config is the kit's, embedded, written for each run: an app's
// own copy, which an agent could weaken, is never read.
func lintFor(goos, pattern string) func(context.Context, Options) Result {
	return func(ctx context.Context, o Options) (r Result) {
		config, err := writeConfig()
		if err != nil {
			return failed(err)
		}
		defer func() {
			if err := os.Remove(config); err != nil && r.Err == nil {
				r.Err = fmt.Errorf("remove the lint config: %w", err)
			}
		}()
		out, err := run.Run(ctx, run.Cmd{Dir: o.Dir, Env: []string{"GOOS=" + goos}, Argv: []string{
			"golangci-lint", "run", "--config", config, "--output.text.print-issued-lines=false", "--show-stats=false", pattern,
		}})
		return ran{out, err}.result(1, out.Stdout)
	}
}

func writeConfig() (string, error) {
	f, err := os.CreateTemp("", "kit-golangci-*.yml")
	if err != nil {
		return "", fmt.Errorf("write the lint config: %w", err)
	}
	_, werr := f.Write(lint.Config)
	if err := errors.Join(werr, f.Close()); err != nil {
		return "", fmt.Errorf("write the lint config: %w", err)
	}
	return f.Name(), nil
}

func vetFor(goos, pattern string) func(context.Context, Options) Result {
	return func(_ context.Context, o Options) Result {
		cfg := &packages.Config{Mode: packages.LoadAllSyntax, Dir: o.Dir, Tests: true, Env: append(os.Environ(), "GOOS="+goos)}
		pkgs, err := packages.Load(cfg, pattern)
		if err != nil {
			return failed(fmt.Errorf("load packages: %w", err))
		}
		graph, err := checker.Analyze(all.Analyzers(), pkgs, nil)
		if err != nil {
			return failed(fmt.Errorf("analyze: %w", err))
		}
		var lines []string
		for act := range graph.All() {
			lines = append(lines, diagnostics(o.Dir, act)...)
		}
		return found(lines)
	}
}

func diagnostics(dir string, act *checker.Action) []string {
	if !act.IsRoot {
		return nil
	}
	var lines []string
	for _, d := range act.Diagnostics {
		p := act.Package.Fset.Position(d.Pos)
		lines = append(lines, position(dir, at{p.Filename, p.Line}, act.Analyzer.Name+": "+d.Message))
	}
	return lines
}

// WHY: archtest and sweep both list what they find with a file, a line, a
// rule and a message; one gate prints either.
func listGate[T any](check func(string) ([]T, error), line func(T) (at, string)) func(context.Context, Options) Result {
	return func(_ context.Context, o Options) Result {
		fs, err := check(o.Dir)
		if err != nil {
			return failed(err)
		}
		lines := make([]string, 0, len(fs))
		for _, f := range fs {
			p, msg := line(f)
			lines = append(lines, position(o.Dir, p, msg))
		}
		return found(lines)
	}
}

func archLine(v archtest.Violation) (at, string) {
	return at{v.File, v.Line}, v.Rule + ": " + v.Message
}

func sweepLine(f sweep.Finding) (at, string) { return at{f.File, f.Line}, f.Rule + ": " + f.Message }

func guardGate(ctx context.Context, o Options) Result {
	_, all, err := guard.Run(ctx, o.Dir, o.Base)
	if err != nil {
		return failed(err)
	}
	trees, err := guard.ReadTrees(o.ApprovedTrees)
	if err != nil {
		return failed(err)
	}
	covered, err := guard.Covered(ctx, all, guard.Approval{Dir: o.Dir, Head: o.Approved, Trees: trees})
	if err != nil {
		return failed(err)
	}
	blocking := guard.Blocking(all, o.CI, covered)
	r := Result{}
	for _, f := range all {
		line := fmt.Sprintf("%s:%d: %s: %s", f.File, f.Line, f.Rule, f.Message)
		if slices.Contains(blocking, f) {
			r.Findings = append(r.Findings, line)
		} else {
			r.Notes = append(r.Notes, line)
		}
	}
	return r
}

func nonEmpty(text []byte) []string {
	var lines []string
	for l := range strings.Lines(string(text)) {
		if l = strings.TrimRight(l, "\r\n"); l != "" && !strings.HasPrefix(l, " ") && !strings.HasPrefix(l, "\t") {
			lines = append(lines, l)
		}
	}
	return lines
}

func specGate(ctx context.Context, o Options) Result {
	fs, err := spec.Prove(ctx, o.Dir)
	if err != nil {
		return failed(err)
	}
	return found(spec.Lines(fs))
}
