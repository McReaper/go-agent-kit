// Package check runs every gate of the kit on a module, the same way on an
// agent's machine and in CI: lint and the kit's analyzers per target
// system, archtest, sweep, dead code, nil flows, known vulnerabilities,
// the tests and the guard.
package check

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

// Options says what to check and how strictly.
type Options struct {
	Dir      string
	Base     string
	GOOS     []string
	CI       bool
	Approved bool
	Race     bool
}

// Result is one gate's outcome.
type Result struct {
	Gate     string
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
	name string
	run  func(context.Context, Options) Result
}

// Run runs every gate in turn and returns their results.
func Run(ctx context.Context, o Options) []Result {
	return runGates(ctx, o, gates(o))
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
		r.Gate = g.name
		out = append(out, r)
	}
	return out
}

func perSystem(o Options, pattern string) []gate {
	var gs []gate
	for _, goos := range o.GOOS {
		gs = append(gs, gate{"golangci-lint (" + goos + ")", lintFor(goos, pattern)}, gate{"kitvet (" + goos + ")", vetFor(goos, pattern)})
	}
	return gs
}

func gates(o Options) []gate {
	gs := perSystem(o, "./...")
	return append(gs,
		gate{"archtest", archGate}, gate{"sweep", sweepGate}, gate{"deadcode", deadcodeGate},
		gate{"nilaway", nilawayGate}, gate{"govulncheck", vulnGate}, gate{"go test", testGate}, gate{"guard", guardGate})
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

func position(dir, file string, line int, msg string) string {
	return fmt.Sprintf("%s:%d: %s", rel(dir, file), line, msg)
}
