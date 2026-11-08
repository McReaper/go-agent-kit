package check

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/McReaper/go-agent-kit/internal/platform/git"
	"github.com/McReaper/go-agent-kit/internal/platform/gitleaks"
	"github.com/McReaper/go-agent-kit/internal/platform/run"
	"github.com/McReaper/go-agent-kit/kitinit"
)

// WHY: workflows, scripts, secrets and the dependency bot's config aren't
// Go: the Go gates never read them, and an app's CI runs kit check alone.
func fileGates() []gate {
	return []gate{
		{Gate{"actionlint", "actionlint", "Workflows", "The workflows are valid."}, actionlintGate},
		{Gate{"zizmor", "zizmor", "Workflow security", "The workflows have no known security flaw."}, zizmorGate},
		{Gate{"shellcheck", "shellcheck", "Shell scripts", "The shell scripts have no known pitfall."}, shellcheckGate},
		{Gate{"psscriptanalyzer", "psscriptanalyzer", "PowerShell scripts", "The PowerShell scripts have no known pitfall."}, psscriptanalyzerGate},
		{Gate{"gitleaks", "gitleaks", "Secrets", "No secret is committed."}, gitleaksGate},
		{Gate{"kit-files", "kit files", "Kit files", "The kit's workflows are the ones this kit version writes."}, kitFilesGate},
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
	out, err := run.Command(ctx, o.Dir, append([]string{"actionlint", "-oneline"}, files...)...)
	return ran{out, err}.result(1, out.Stdout)
}

var zizmorLine = regexp.MustCompile(`^::\w+ file=([^,]+),line=(\d+),title=([^:]+)::[^:]*:\d+: (.*)$`)

// WHY: zizmor exits 11 to 14 by the worst finding's severity.
func zizmorGate(ctx context.Context, o Options) Result {
	files, err := workflows(ctx, o.Dir)
	if err != nil || len(files) == 0 {
		return Result{Err: err}
	}
	out, err := run.Command(ctx, o.Dir, append([]string{"zizmor", "--offline", "--format", "github"}, files...)...)
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
	out, err := run.Command(ctx, o.Dir, append([]string{"shellcheck", "-f", "gcc"}, scripts...)...)
	return ran{out, err}.result(1, out.Stdout)
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

// WHY: the kit's workflows are the files an app keeps; once a kit version
// needs them otherwise, the app hears it from its own check.
func kitFilesGate(_ context.Context, o Options) Result {
	current, err := os.ReadFile(filepath.Join(o.Dir, ".github", "workflows", "kit.yml"))
	if errors.Is(err, fs.ErrNotExist) {
		return Result{}
	}
	if err != nil {
		return failed(err)
	}
	want, err := kitinit.WorkflowsFor(current)
	if err != nil {
		return found([]string{".github/workflows/kit.yml:1: kit-files: T2: " + err.Error()})
	}
	var lines []string
	for _, f := range want {
		have, err := os.ReadFile(filepath.Join(o.Dir, filepath.FromSlash(f.Path)))
		switch {
		case errors.Is(err, fs.ErrNotExist):
			lines = append(lines, f.Path+":1: kit-files: T2: this kit version writes it and the repository has none: run kit init, which writes it")
		case err != nil:
			return failed(err)
		case !bytes.Equal(have, f.Content):
			lines = append(lines, f.Path+":1: kit-files: T2: it differs from the one this kit version writes: delete it and run kit init, which writes it again")
		}
	}
	return found(lines)
}
