package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/McReaper/rumblestrip-go/check"
	"github.com/McReaper/rumblestrip-go/mutate"
	"github.com/spf13/cobra"
)

var errScoreRange = errors.New("--min is a share, from 0 to 1")

type mutateOptions struct {
	base    string
	min     float64
	github  bool
	summary string
}

func mutateCommand(out io.Writer) *cobra.Command {
	var o mutateOptions
	cmd := leaf("mutate", "Mutate the lines a change touched; fail when the tests kill too few mutants", func(cmd *cobra.Command) error { return runMutate(cmd.Context(), out, o) })
	cmd.Flags().StringVar(&o.base, "base", "origin/main", "the branch the change goes into")
	cmd.Flags().Float64Var(&o.min, "min", 0.8, "the share of the mutants the tests must kill")
	cmd.Flags().BoolVar(&o.github, "github", false, "also annotate the lines whose mutants lived (GitHub Actions)")
	cmd.Flags().StringVar(&o.summary, "summary", "", "with --github, leave the check's row of the run's summary in this directory, and its section when it fails")
	return cmd
}

func runMutate(ctx context.Context, out io.Writer, o mutateOptions) error {
	if o.min < 0 || o.min > 1 {
		return errScoreRange
	}
	dir, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("mutate: %w", err)
	}
	r, err := mutate.Run(ctx, dir, o.base)
	if err != nil {
		return fmt.Errorf("mutate: %w", err)
	}
	if o.github {
		if err := reportGitHub(out, o.summary, []check.Result{mutationResult(r, o)}); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(out, mutateReport(r, o)); err != nil {
		return fmt.Errorf("print the mutants: %w", err)
	}
	if r.Score < o.min {
		return errFindings
	}
	return nil
}

// WHY: on GitHub, mutation is one more check: a lived mutant is a finding
// at its line, which the owner reads as the change the tests missed.
func mutationResult(r mutate.Report, o mutateOptions) check.Result {
	res := check.Result{
		ID: "mutate", Gate: "mutate", Title: "Mutation",
		Purpose: fmt.Sprintf("The tests catch the changes made to the lines the pull request added: %.0f%% of them at least.", 100*o.min),
	}
	if r.Score >= o.min {
		return res
	}
	for _, m := range r.Mutants {
		if m.Killed {
			continue
		}
		change := "removes this code"
		if m.Wrote != "" {
			change = "makes it " + "`" + m.Wrote + "`"
		}
		res.Findings = append(res.Findings, fmt.Sprintf("%s:%d: the tests still pass when a change %s (%s)", m.File, m.Line, change, m.Mutator))
	}
	return res
}

func mutateReport(r mutate.Report, o mutateOptions) string {
	var b strings.Builder
	killed := 0
	for _, m := range r.Mutants {
		if m.Killed {
			killed++
			continue
		}
		wrote := m.Wrote
		if wrote == "" {
			wrote = "(code removed)"
		}
		fmt.Fprintf(&b, "%s:%d: lived: %s: %s\n", m.File, m.Line, m.Mutator, wrote)
	}
	fmt.Fprintf(&b, "rumblestrip mutate: the tests killed %d of the %d mutants of the lines changed since %s: %.0f%%, %.0f%% needed\n",
		killed, len(r.Mutants), o.base, 100*r.Score, 100*o.min)
	if r.ErrorReturns > 0 {
		fmt.Fprintf(&b, "rumblestrip mutate: %d more only drop an error's return, not counted: the compiler and errcheck refuse a dropped error\n", r.ErrorReturns)
	}
	return b.String()
}
