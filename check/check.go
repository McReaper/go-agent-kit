// Package check runs every gate of the kit on a module, the same way on an
// agent's machine and in CI: lint and the kit's analyzers per target
// system, archtest, sweep, dead code, nil flows, known vulnerabilities,
// the tests, the spec's proof, the guard, and the workflows, scripts and
// secrets.
package check

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/McReaper/go-agent-kit/archtest"
	"github.com/McReaper/go-agent-kit/sweep"
)

// Options says what to check and how strictly.
type Options struct {
	Dir  string
	Base string
	GOOS []string
	Race bool
	// Title is the pull request's title, which only CI has: its gate runs
	// when it is set.
	Title string
	// Only names the gates to run by their ID; empty runs every one.
	Only []string
}

// Result is one gate's outcome: Gate is its name in the log, Title and
// Purpose what the owner reads.
type Result struct {
	ID       string
	Gate     string
	Title    string
	Purpose  string
	Findings []string
	Notes    []string
	Err      error
}

// Failed reports whether the gate fails the check: it found something, or
// it couldn't run. Notes, the changes for the owner to approve, never fail.
func (r Result) Failed() bool {
	return r.Err != nil || len(r.Findings) > 0
}

type gate struct {
	Gate
	run func(context.Context, Options) Result
}

// Gate is a gate as CI and the owner name it: ID for --only and the CI
// job that runs it alone, Name in the log, Title and Purpose for the owner.
type Gate struct {
	ID, Name, Title, Purpose string
}

// Run runs every gate in turn, or the ones o.Only names, and returns their
// results.
func Run(ctx context.Context, o Options) ([]Result, error) {
	gs := gates(o)
	if len(o.Only) > 0 {
		var err error
		if gs, err = only(gs, o.Only); err != nil {
			return nil, err
		}
	}
	return runGates(ctx, o, gs), nil
}

// Gates lists the gates Run runs for o, in order.
func Gates(o Options) []Gate {
	var out []Gate
	for _, g := range gates(o) {
		out = append(out, g.Gate)
	}
	return out
}

func only(gs []gate, ids []string) ([]gate, error) {
	var out []gate
	for _, id := range ids {
		i := slices.IndexFunc(gs, func(g gate) bool { return g.ID == id })
		if i < 0 {
			return nil, fmt.Errorf("kit check: no gate %q; the gates are %s (title needs --title)", id, strings.Join(gateIDs(gs), ", "))
		}
		out = append(out, gs[i])
	}
	return out, nil
}

func gateIDs(gs []gate) []string {
	ids := make([]string, 0, len(gs))
	for _, g := range gs {
		ids = append(ids, g.ID)
	}
	return ids
}

// Package runs the gates that judge one package on its own, the linters
// and the kit's analyzers per GOOS: what an agent hears after each edit.
// pattern is a package path relative to o.Dir, like ./internal/notes.
func Package(ctx context.Context, o Options, pattern string) []Result {
	return runGates(ctx, o, perSystem(o, pattern))
}

func runGates(ctx context.Context, o Options, gs []gate) []Result {
	var out []Result
	for _, g := range gs {
		r := g.run(ctx, o)
		r.ID, r.Gate, r.Title, r.Purpose = g.ID, g.Name, g.Title, g.Purpose
		out = append(out, r)
	}
	return out
}

func perSystem(o Options, pattern string) []gate {
	var gs []gate
	for _, goos := range o.GOOS {
		system := systemName(goos)
		gs = append(gs,
			gate{Gate{"golangci-lint-" + goos, "golangci-lint (" + goos + ")", "Lint (" + system + ")", "The code follows the kit's lint rules, built for " + system + "."}, lintFor(goos, pattern)},
			gate{Gate{"kitvet-" + goos, "kitvet (" + goos + ")", "Kit analyzers (" + system + ")", "The code follows the kit's own rules, built for " + system + "."}, vetFor(goos, pattern)})
	}
	return gs
}

func systemName(goos string) string {
	switch goos {
	case "linux":
		return "Linux"
	case "windows":
		return "Windows"
	case "darwin":
		return "macOS"
	}
	return goos
}

func gates(o Options) []gate {
	gs := perSystem(o, "./...")
	gs = append(gs,
		gate{Gate{"archtest", "archtest", "Architecture", "Each module reaches the others through their contract only."}, listGate(archtest.Check, archLine)},
		gate{Gate{"sweep", "sweep", "Sweep", "No code only tests reach, no name exported for its own package, no log message written twice."}, listGate(sweep.Check, sweepLine)},
		gate{Gate{"deadcode", "deadcode", "Dead code", "No function the program never calls."}, deadcodeGate},
		gate{Gate{"nilaway", "nilaway", "Nil safety", "No nil value reaches a place that uses it."}, nilawayGate},
		gate{Gate{"govulncheck", "govulncheck", "Known vulnerabilities", "The code calls no function with a known vulnerability."}, vulnGate},
		gate{Gate{"go-test", "go test", "Tests", "Every test passes."}, testGate},
		gate{Gate{"spec", "spec", "Spec", "Each built scenario passes, and fails once a value it expects is altered."}, specGate},
		gate{Gate{"guard", "guard", "Guard", "No check is switched off: no //nolint, no skipped test, no build tag that hides a file."}, guardGate})
	gs = append(gs, fileGates()...)
	if o.Title != "" {
		gs = append(gs, gate{Gate{"title", "title", "Title", "The pull request's title is a conventional commit of " + strconv.Itoa(MaxTitle) + " characters at most."}, titleGate})
	}
	return gs
}

func failed(err error) Result {
	return Result{Err: err}
}

func found(lines []string) Result {
	slices.Sort(lines)
	return Result{Findings: slices.Compact(lines)}
}

func rel(dir, file string) string {
	if r, err := filepath.Rel(dir, file); err == nil && !strings.HasPrefix(r, "..") {
		return filepath.ToSlash(r)
	}
	return file
}

type at struct {
	file string
	line int
}

func position(dir string, p at, msg string) string {
	return fmt.Sprintf("%s:%d: %s", rel(dir, p.file), p.line, msg)
}
