// Command kit runs the kit's checks on a repository.
//
//	kit guard --base origin/main [--ci [--approved]] [--json]
//	kit check [--base origin/main] [--ci [--approved]]
//	kit sweep [--json]
//	kit mutate [--base origin/main] [--min 0.8]
//	kit review [--base origin/main] [--labels | --json]
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/McReaper/go-agent-kit/guard"
	"github.com/spf13/cobra"
)

var errFindings = errors.New("the change breaks the kit's rules")

func main() {
	err := root(os.Stdout).ExecuteContext(context.Background())
	if code, ok := errors.AsType[exitCode](err); ok {
		os.Exit(int(code))
	}
	if err != nil {
		if !errors.Is(err, errFindings) {
			fmt.Fprintln(os.Stderr, "kit:", err)
		}
		os.Exit(1)
	}
}

func root(out io.Writer) *cobra.Command {
	r := &cobra.Command{Use: "kit", Short: "The anti-drift kit's checks", SilenceUsage: true, SilenceErrors: true}
	r.AddCommand(checkCommand(out), guardCommand(out), sweepCommand(out), mutateCommand(out), reviewCommand(out), appCommand(out), hookCommand())
	return r
}

type guardOptions struct {
	base     string
	ci       bool
	approved bool
	asJSON   bool
}

func guardCommand(out io.Writer) *cobra.Command {
	var o guardOptions
	cmd := &cobra.Command{
		Use:   "guard",
		Short: "Check a change for bypasses, protected files and size",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runGuard(cmd.Context(), out, o)
		},
	}
	cmd.Flags().StringVar(&o.base, "base", "origin/main", "the branch the change goes into")
	cmd.Flags().BoolVar(&o.ci, "ci", false, "the merge check: changes that need approval fail until --approved")
	cmd.Flags().BoolVar(&o.approved, "approved", false, "a code owner approved the change's latest commit")
	cmd.Flags().BoolVar(&o.asJSON, "json", false, "print the findings as JSON")
	return cmd
}

func runGuard(ctx context.Context, out io.Writer, o guardOptions) error {
	d, err := guard.Read(ctx, ".", o.base)
	if err != nil {
		return err
	}
	findings := guard.Check(d)
	blocking := guard.Blocking(findings, o.ci, o.approved)
	if o.asJSON {
		err = writeJSON(out, findings)
	} else {
		err = report(out, findings, len(d.Files), o.base)
	}
	if err != nil {
		return err
	}
	if len(blocking) > 0 {
		return errFindings
	}
	return nil
}

func report(out io.Writer, findings []guard.Finding, files int, base string) error {
	var broken, approval []guard.Finding
	for _, f := range findings {
		if f.NeedsApproval() {
			approval = append(approval, f)
		} else {
			broken = append(broken, f)
		}
	}
	var b strings.Builder
	section(&b, "Blocking: fix these", broken)
	section(&b, "Needs the owner's approval before merge (CI holds the change until then)", approval)
	fmt.Fprintf(&b, "kit guard: %d files changed since %s (commits, edits and new files): %d blocking, %d for the owner to approve\n",
		files, base, len(broken), len(approval))
	if _, err := io.WriteString(out, b.String()); err != nil {
		return fmt.Errorf("print findings: %w", err)
	}
	return nil
}

func section(b *strings.Builder, title string, fs []guard.Finding) {
	if len(fs) == 0 {
		return
	}
	fmt.Fprintf(b, "%s:\n", title)
	for _, f := range fs {
		fmt.Fprintf(b, "  %s:%d: %s: %s\n", f.File, f.Line, f.Rule, f.Message)
	}
}

func writeJSON(out io.Writer, findings []guard.Finding) error {
	type row struct {
		Rule          string `json:"rule"`
		File          string `json:"file"`
		Line          int    `json:"line"`
		Message       string `json:"message"`
		NeedsApproval bool   `json:"needs_approval"`
	}
	rows := make([]row, 0, len(findings))
	for _, f := range findings {
		rows = append(rows, row{Rule: f.Rule, File: f.File, Line: f.Line, Message: f.Message, NeedsApproval: f.NeedsApproval()})
	}
	if err := json.NewEncoder(out).Encode(rows); err != nil {
		return fmt.Errorf("print findings: %w", err)
	}
	return nil
}
