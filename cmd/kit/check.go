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
	r := checkRun{o: check.Options{Base: "origin/main", GOOS: []string{"linux", "windows"}, Race: runtime.GOOS == "linux"}}
	o := &r.o
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Run every gate of the kit on the module: the same command on your machine and in CI",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runCheck(cmd, out, r)
		},
	}
	cmd.Flags().StringVar(&o.Base, "base", o.Base, "the branch the change goes into, for the guard")
	cmd.Flags().StringSliceVar(&o.GOOS, "goos", o.GOOS, "the systems the module builds for")
	cmd.Flags().BoolVar(&o.Race, "race", o.Race, "run the tests with the race detector (needs cgo)")
	cmd.Flags().StringVar(&o.Title, "title", "", "the pull request's title, to check it too (CI)")
	cmd.Flags().StringSliceVar(&o.Only, "only", nil, "run only these gates, by ID (CI runs each in its own job)")
	cmd.Flags().BoolVar(&r.github, "github", false, "also annotate the findings' lines (GitHub Actions)")
	cmd.Flags().StringVar(&r.summary, "summary", "", "with --github, leave each gate's row of the run's summary in this directory, and its section when it fails")
	return cmd
}

// WHY: --github changes how the results are written, not what is checked.
type checkRun struct {
	o       check.Options
	github  bool
	summary string
}

func runCheck(cmd *cobra.Command, out io.Writer, r checkRun) error {
	dir, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("check: %w", err)
	}
	r.o.Dir = dir
	results, err := check.Run(cmd.Context(), r.o)
	if err != nil {
		return err
	}
	if r.github {
		if err := reportGitHub(out, r.summary, results); err != nil {
			return err
		}
	}
	return reportCheck(out, results)
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
	switch {
	case len(failedGates) == 0:
		b.WriteString("kit check: every gate passes\n")
	case len(failedGates) == 1:
		fmt.Fprintf(&b, "kit check: 1 gate fails: %s\n", failedGates[0])
	default:
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
