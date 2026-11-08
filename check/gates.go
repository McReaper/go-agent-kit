package check

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/McReaper/rumblestrip-go/analyzers/all"
	"github.com/McReaper/rumblestrip-go/archtest"
	"github.com/McReaper/rumblestrip-go/guard"
	"github.com/McReaper/rumblestrip-go/internal/platform/run"
	"github.com/McReaper/rumblestrip-go/lint"
	"github.com/McReaper/rumblestrip-go/spec"
	"github.com/McReaper/rumblestrip-go/sweep"
	"golang.org/x/tools/go/analysis/checker"
	"golang.org/x/tools/go/packages"
)

// WHY: the config is Rumblestrip's, embedded, written for each run: an app's
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
	f, err := os.CreateTemp("", "rumblestrip-golangci-*.yml")
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

// WHY: what is the owner's to approve never fails the check, here or in
// CI: the review sheet's job holds the merge for it, through the owner's
// approval status, from a job that runs none of the change's code.
func guardGate(ctx context.Context, o Options) Result {
	_, all, err := guard.Run(ctx, o.Dir, o.Base)
	if err != nil {
		return failed(err)
	}
	r := Result{}
	for _, f := range all {
		line := fmt.Sprintf("%s:%d: %s: %s", f.File, f.Line, f.Rule, f.Message)
		if f.NeedsApproval() {
			r.Notes = append(r.Notes, line)
		} else {
			r.Findings = append(r.Findings, line)
		}
	}
	return r
}

// MaxTitle is the longest pull request title: GitHub cuts a longer one in
// its lists, and the title becomes the commit's subject on the default
// branch.
const MaxTitle = 72

var conventional = regexp.MustCompile(`^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert|spec)(\([a-z0-9][a-z0-9/._-]*\))?!?: \S`)

func titleGate(_ context.Context, o Options) Result {
	var lines []string
	if n := utf8.RuneCountInString(o.Title); n > MaxTitle {
		lines = append(lines, fmt.Sprintf("title: the title has %d characters, the limit is %d: say less, the description holds the rest", n, MaxTitle))
	}
	if !conventional.MatchString(o.Title) {
		lines = append(lines, "title: the title is no conventional commit: write type(scope): what changes, the type one of feat, fix, docs, refactor, perf, test, build, ci, chore, revert, style, or spec for a spec's pull request")
	}
	return Result{Findings: lines}
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
