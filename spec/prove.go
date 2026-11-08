package spec

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"

	"github.com/McReaper/go-agent-kit/internal/platform/modcopy"
	"github.com/McReaper/go-agent-kit/internal/platform/run"
)

// Finding is a scenario that proves less than it says.
type Finding struct {
	Rule    string
	File    string
	Line    int
	Message string
}

// Prove checks that each built scenario can fail. Run alone, it must pass
// as written; run once more for each value its Then steps expect, that
// value altered, it must fail. A step that passes whatever it is told to
// expect checks nothing, and the owner who approved the scenario would be
// trusting a sentence. dir is the root of the module and of its git
// repository; the runs happen on a copy.
func Prove(ctx context.Context, dir string) (out []Finding, err error) {
	scenarios, err := Read(dir)
	if err != nil {
		return nil, err
	}
	var jobs []job
	for _, s := range scenarios {
		if !s.WIP {
			fs, js := plan(s)
			out, jobs = append(out, fs...), append(jobs, js...)
		}
	}
	if len(jobs) == 0 {
		return out, nil
	}
	work, err := modcopy.Copy(ctx, dir)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(work)) }()
	bin, err := buildSpec(ctx, work)
	if err != nil {
		return nil, err
	}
	if bin == "" {
		return append(out, Finding{
			"spec-untested", Dir, 0,
			"T2: features/ holds the spec but no test that runs it: add the godog test that runs every feature file of the directory",
		}), nil
	}
	return append(out, verdicts(jobs, runner{work, bin}.all(ctx, jobs))...), nil
}

// WHY: a job runs one scenario, as written (step -1), or with the value one
// of its Then steps expects altered, so a verdict knows which.
type job struct {
	run   Scenario
	of    Scenario
	step  int
	value string
}

func plan(s Scenario) ([]Finding, []job) {
	var out []Finding
	jobs := []job{{run: s, of: s, step: -1}}
	for i, st := range s.Steps {
		if st.Kind != Then {
			continue
		}
		vs := values(st)
		if len(vs) == 0 {
			out = append(out, Finding{"spec-novalue", s.File, st.Line, fmt.Sprintf(
				"T2: `%s` expects no value the kit can alter, so nothing shows it can fail: say what it expects, a quoted text or a number (Then it is refused with \"the limit is 80\")", "Then "+st.Text)})
		}
		for _, v := range vs {
			m := s
			m.Steps = append([]Step(nil), s.Steps...)
			m.Steps[i] = v.alter()
			jobs = append(jobs, job{run: m, of: s, step: i, value: v.raw})
		}
	}
	return out, jobs
}

type outcome struct {
	passed    bool
	undefined bool
	reason    string
	err       error
}

func buildSpec(ctx context.Context, work string) (string, error) {
	bin := filepath.Join(work, "kit-spec.test")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	out, err := run.Command(ctx, work, "go", "test", "-c", "-o", bin, "./"+Dir)
	if err != nil {
		return "", err
	}
	// WHY: with no Go file left in features/, go test has no package to build.
	if out.Code != 0 && !strings.Contains(string(out.Stderr), "no Go files") {
		return "", fmt.Errorf("build the spec's test: %s%s", out.Stdout, out.Stderr)
	}
	if _, err := os.Stat(bin); errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	return bin, nil
}

// WHY: each run gets a directory of its own beside features/, holding
// features/'s other files and the one scenario, so the test finds what it
// reads at the same relative paths and the runs don't see each other.
// WHY: the runs share the module's copy and the spec's test, built once.
type runner struct {
	work, bin string
}

func (r runner) all(ctx context.Context, jobs []job) []outcome {
	out := make([]outcome, len(jobs))
	next := make(chan int)
	var wg sync.WaitGroup
	for range max(1, runtime.NumCPU()/2) {
		wg.Go(func() {
			for i := range next {
				out[i] = r.one(ctx, filepath.Join(r.work, fmt.Sprintf(".kit-spec-%d", i)), jobs[i].run)
			}
		})
	}
	for i := range jobs {
		next <- i
	}
	close(next)
	wg.Wait()
	return out
}

var (
	ansi      = regexp.MustCompile("\x1b\\[[0-9;]*m")
	undefined = regexp.MustCompile(`\b\d+ undefined\b`)
	stepError = regexp.MustCompile(`(?m)^\s+Error: (.+)$`)
)

// WHY: godog lists each failed step again under "--- Failed steps:", with
// its error: the first one is why the scenario failed.
func failedStep(text string) string {
	_, failed, ok := strings.Cut(text, "--- Failed steps:")
	if !ok {
		return ""
	}
	if m := stepError.FindStringSubmatch(failed); m != nil {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func (r runner) one(ctx context.Context, dir string, s Scenario) outcome {
	if err := os.CopyFS(dir, os.DirFS(filepath.Join(filepath.Dir(dir), Dir))); err != nil {
		return outcome{err: fmt.Errorf("prepare a run of %q: %w", s.Name, err)}
	}
	if err := removeFeatures(dir); err != nil {
		return outcome{err: err}
	}
	if err := os.WriteFile(filepath.Join(dir, "scenario.feature"), []byte(s.Gherkin()), 0o600); err != nil {
		return outcome{err: fmt.Errorf("prepare a run of %q: %w", s.Name, err)}
	}
	res, err := run.Command(ctx, dir, r.bin, "-test.timeout=5m")
	if err != nil {
		return outcome{err: err}
	}
	text := ansi.ReplaceAllString(string(res.Stdout)+string(res.Stderr), "")
	return outcome{passed: res.Code == 0, undefined: res.Code != 0 && undefined.MatchString(text), reason: failedStep(text)}
}

func removeFeatures(dir string) (err error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return fmt.Errorf("prepare a run: %w", err)
	}
	defer func() { err = errors.Join(err, root.Close()) }()
	return fs.WalkDir(root.FS(), ".", func(p string, e fs.DirEntry, err error) error {
		if err != nil || e.IsDir() || path.Ext(p) != ".feature" {
			return err
		}
		return root.Remove(p)
	})
}

func verdicts(jobs []job, outs []outcome) []Finding {
	var out []Finding
	alone := map[string]bool{}
	for i, j := range jobs {
		o := outs[i]
		key := fmt.Sprintf("%s:%d", j.of.File, j.of.Line)
		switch {
		case o.err != nil:
			out = append(out, Finding{"spec-run", j.of.File, j.of.Line, "T2: " + o.err.Error()})
		case j.step < 0 && !o.passed:
			alone[key] = true
			out = append(out, Finding{"spec-alone", j.of.File, j.of.Line, aloneMessage(j.of.Name, o.reason)})
		case j.step < 0 || alone[key]:
		case o.passed:
			out = append(out, Finding{"spec-vacuous", j.of.File, j.of.Steps[j.step].Line, fmt.Sprintf(
				"T2: `%s` still passes when it expects %q instead of %q: the step checks nothing of that value; make it fail when the value differs",
				"Then "+j.of.Steps[j.step].Text, altered(j.value), j.value)})
		case o.undefined:
			out = append(out, Finding{"spec-pattern", j.of.File, j.of.Steps[j.step].Line, fmt.Sprintf(
				"T2: with %q instead of %q, `%s` matches no step definition, so the kit can't show the step checks it: match the value with a pattern that takes any value, like \"([^\"]*)\" or (\\d+)",
				altered(j.value), j.value, "Then "+j.of.Steps[j.step].Text)})
		}
	}
	return out
}

// WHY: a scenario fails alone either because the program doesn't do what it
// says yet, or because it leans on what another scenario left behind; the
// step's error tells which, so it comes first.
func aloneMessage(name, reason string) string {
	if reason == "" {
		reason = "godog named no failed step"
	}
	return fmt.Sprintf("T2: %q fails: %s. Make the program do what it says; if it passes with the other scenarios, make it set up all it needs, since the kit runs each alone", name, reason)
}

// Lines writes findings the way the kit prints them, file:line: rule: message.
func Lines(fs []Finding) []string {
	out := make([]string, 0, len(fs))
	for _, f := range fs {
		out = append(out, fmt.Sprintf("%s:%d: %s: %s", f.File, f.Line, f.Rule, strings.TrimSpace(f.Message)))
	}
	return out
}
