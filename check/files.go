package check

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/McReaper/go-agent-kit/internal/platform/git"
	"github.com/McReaper/go-agent-kit/internal/platform/gitleaks"
	"github.com/McReaper/go-agent-kit/internal/platform/run"
)

// WHY: workflows, scripts, secrets and the dependency bot's config aren't
// Go: the Go gates never read them, and an app's CI runs kit check alone.
func fileGates() []gate {
	return []gate{
		{"actionlint", actionlintGate},
		{"zizmor", zizmorGate},
		{"shellcheck", shellcheckGate},
		{"psscriptanalyzer", psscriptanalyzerGate},
		{"gitleaks", gitleaksGate},
		{"renovate", renovateGate},
	}
}

func tracked(ctx context.Context, dir string) ([]string, error) {
	out, err := git.Run(ctx, dir, "ls-files", "-z", "--cached", "--others", "--exclude-standard")
	if err != nil {
		return nil, err
	}
	var files []string
	for f := range strings.SplitSeq(strings.TrimSuffix(out, "\x00"), "\x00") {
		if f != "" {
			files = append(files, f)
		}
	}
	return files, nil
}

func workflows(ctx context.Context, dir string) ([]string, error) {
	files, err := tracked(ctx, dir)
	return slices.DeleteFunc(files, func(f string) bool {
		ext := path.Ext(f)
		return !strings.HasPrefix(f, ".github/workflows/") || ext != ".yml" && ext != ".yaml"
	}), err
}

func actionlintGate(ctx context.Context, o Options) Result {
	files, err := workflows(ctx, o.Dir)
	if err != nil || len(files) == 0 {
		return Result{Err: err}
	}
	out, err := run.Command(ctx, o.Dir, nil, append([]string{"actionlint", "-oneline"}, files...)...)
	return fromLines(out, err, 1, out.Stdout)
}

var zizmorLine = regexp.MustCompile(`^::\w+ file=([^,]+),line=(\d+),title=([^:]+)::[^:]*:\d+: (.*)$`)

// WHY: zizmor exits 11 to 14 by the worst finding's severity.
func zizmorGate(ctx context.Context, o Options) Result {
	files, err := workflows(ctx, o.Dir)
	if err != nil || len(files) == 0 {
		return Result{Err: err}
	}
	out, err := run.Command(ctx, o.Dir, nil, append([]string{"zizmor", "--offline", "--format", "github"}, files...)...)
	if err != nil {
		return failed(err)
	}
	if out.Code != 0 && (out.Code < 11 || out.Code > 14) {
		return failed(fmt.Errorf("exit %d: %s", out.Code, strings.TrimSpace(string(out.Stderr))))
	}
	var lines []string
	for l := range strings.Lines(string(out.Stdout)) {
		if m := zizmorLine.FindStringSubmatch(strings.TrimSpace(l)); m != nil {
			lines = append(lines, fmt.Sprintf("%s:%s: %s: %s", m[1], m[2], m[3], m[4]))
		}
	}
	return found(lines)
}

func shellcheckGate(ctx context.Context, o Options) Result {
	files, err := tracked(ctx, o.Dir)
	if err != nil {
		return failed(err)
	}
	scripts, err := shellScripts(o.Dir, files)
	if err != nil || len(scripts) == 0 {
		return Result{Err: err}
	}
	out, err := run.Command(ctx, o.Dir, nil, append([]string{"shellcheck", "-f", "gcc"}, scripts...)...)
	return fromLines(out, err, 1, out.Stdout)
}

// WHY: a git hook has no extension; its first line says it is a script.
func shellScripts(dir string, files []string) (scripts []string, err error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", dir, err)
	}
	defer func() { err = errors.Join(err, root.Close()) }()
	for _, f := range files {
		if path.Ext(f) == ".sh" || path.Ext(f) == "" && hasShellShebang(root, f) {
			scripts = append(scripts, f)
		}
	}
	return scripts, nil
}

var shellShebang = regexp.MustCompile(`^#!\s*(/bin/|/usr/bin/env\s+)(ba|da|k)?sh\b`)

func hasShellShebang(root *os.Root, f string) bool {
	file, err := root.Open(f)
	if err != nil {
		return false
	}
	head := make([]byte, 32)
	n, rerr := file.Read(head)
	if err := errors.Join(file.Close(), ignoreEOF(rerr)); err != nil {
		return false
	}
	first, _, _ := bytes.Cut(head[:n], []byte("\n"))
	return shellShebang.Match(first)
}

func ignoreEOF(err error) error {
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}

// WHY: the change's commits, then what isn't committed yet, staged or not:
// the agent runs kit check before it commits.
func gitleaksGate(ctx context.Context, o Options) Result {
	var lines []string
	for _, scope := range [][]string{{"--log-opts=" + o.Base + "..HEAD"}, {"--pre-commit"}, {"--pre-commit", "--staged"}} {
		leaks, err := gitleaks.Scan(ctx, o.Dir, scope...)
		if err != nil {
			return failed(err)
		}
		for _, l := range leaks {
			lines = append(lines, fmt.Sprintf("%s:%d: %s: %s", l.File, l.Line, l.Rule, l.Description))
		}
	}
	return found(lines)
}
