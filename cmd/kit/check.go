package main

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/McReaper/go-agent-kit/check"
	"github.com/spf13/cobra"
)

func checkCommand(out io.Writer) *cobra.Command {
	o := check.Options{Base: "origin/main", GOOS: []string{"linux", "windows"}, Race: runtime.GOOS == "linux"}
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Run every gate of the kit on the module: the same command on your machine and in CI",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dir, err := filepath.Abs(".")
			if err != nil {
				return fmt.Errorf("check: %w", err)
			}
			o.Dir = dir
			return reportCheck(out, check.Run(cmd.Context(), o))
		},
	}
	cmd.Flags().StringVar(&o.Base, "base", o.Base, "the branch the change goes into, for the guard")
	cmd.Flags().StringSliceVar(&o.GOOS, "goos", o.GOOS, "the systems the module builds for")
	cmd.Flags().BoolVar(&o.CI, "ci", false, "the merge check: changes that need approval fail until --approved")
	cmd.Flags().BoolVar(&o.Approved, "approved", false, "a code owner approved the change's latest commit")
	cmd.Flags().StringVar(&o.ApprovedTrees, "approved-trees", "", "a directory of the trees of commits a code owner approved: the protected files they hold unchanged stay approved")
	cmd.Flags().BoolVar(&o.Race, "race", o.Race, "run the tests with the race detector (needs cgo)")
	return cmd
}

func reportCheck(out io.Writer, results []check.Result) error {
	var b strings.Builder
	var failedGates []string
	for _, r := range results {
		writeResult(&b, r)
		if r.Failed() {
			failedGates = append(failedGates, r.Gate)
		}
	}
	if len(failedGates) == 0 {
		b.WriteString("kit check: every gate passes\n")
	} else {
		fmt.Fprintf(&b, "kit check: %d gates fail: %s\n", len(failedGates), strings.Join(failedGates, ", "))
	}
	if _, err := io.WriteString(out, b.String()); err != nil {
		return fmt.Errorf("print the check: %w", err)
	}
	if len(failedGates) > 0 {
		return errFindings
	}
	return nil
}

func writeResult(b *strings.Builder, r check.Result) {
	switch {
	case r.Err != nil:
		fmt.Fprintf(b, "FAIL %s: could not run: %v\n", r.Gate, r.Err)
	case len(r.Findings) > 0:
		fmt.Fprintf(b, "FAIL %s\n", r.Gate)
	default:
		fmt.Fprintf(b, "ok   %s\n", r.Gate)
	}
	for _, f := range r.Findings {
		fmt.Fprintf(b, "  %s\n", f)
	}
	for _, n := range r.Notes {
		fmt.Fprintf(b, "  for the owner to approve before merge: %s\n", n)
	}
}
