package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/McReaper/go-agent-kit/mutate"
	"github.com/spf13/cobra"
)

var errScoreRange = errors.New("--min is a share, from 0 to 1")

type mutateOptions struct {
	base string
	min  float64
}

func mutateCommand(out io.Writer) *cobra.Command {
	var o mutateOptions
	cmd := &cobra.Command{
		Use:   "mutate",
		Short: "Mutate the lines a change touched; fail when the tests kill too few mutants",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runMutate(cmd.Context(), out, o)
		},
	}
	cmd.Flags().StringVar(&o.base, "base", "origin/main", "the branch the change goes into")
	cmd.Flags().Float64Var(&o.min, "min", 0.8, "the share of the mutants the tests must kill")
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
	if _, err := io.WriteString(out, mutateReport(r, o)); err != nil {
		return fmt.Errorf("print the mutants: %w", err)
	}
	if r.Score < o.min {
		return errFindings
	}
	return nil
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
	fmt.Fprintf(&b, "kit mutate: the tests killed %d of the %d mutants of the lines changed since %s: %.0f%%, %.0f%% needed\n",
		killed, len(r.Mutants), o.base, 100*r.Score, 100*o.min)
	if r.ErrorReturns > 0 {
		fmt.Fprintf(&b, "kit mutate: %d more only drop an error's return, not counted: the compiler and errcheck refuse a dropped error\n", r.ErrorReturns)
	}
	return b.String()
}
