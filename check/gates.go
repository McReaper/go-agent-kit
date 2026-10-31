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
		out, err := run.Command(ctx, o.Dir, []string{"GOOS=" + goos}, "golangci-lint", "run", "--config", config,
			"--output.text.print-issued-lines=false", "--show-stats=false", pattern)
		return fromLines(out, err, 1, out.Stdout)
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
		lines = append(lines, position(dir, p.Filename, p.Line, act.Analyzer.Name+": "+d.Message))
	}
	return lines
}

func archGate(_ context.Context, o Options) Result {
	vs, err := archtest.Check(o.Dir)
	if err != nil {
		return failed(err)
	}
	lines := make([]string, 0, len(vs))
	for _, v := range vs {
		lines = append(lines, position(o.Dir, v.File, v.Line, v.Rule+": "+v.Message))
	}
	return found(lines)
}

func sweepGate(_ context.Context, o Options) Result {
	fs, err := sweep.Check(o.Dir)
	if err != nil {
		return failed(err)
	}
	lines := make([]string, 0, len(fs))
	for _, f := range fs {
		lines = append(lines, position(o.Dir, f.File, f.Line, f.Rule+": "+f.Message))
	}
	return found(lines)
}

func guardGate(ctx context.Context, o Options) Result {
	d, err := guard.Read(ctx, o.Dir, o.Base)
	if err != nil {
		return failed(err)
	}
	all := guard.Check(d)
	blocking := guard.Blocking(all, o.CI, o.Approved)
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
