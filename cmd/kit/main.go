// Command kit runs the kit's checks on a repository.
//
//	kit guard --base origin/main [--ci [--approved] [--approved-trees dir]] [--json]
//	kit check [--base origin/main] [--ci [--approved] [--approved-trees dir]]
//	kit sweep [--json]
//	kit mutate [--base origin/main] [--min 0.8]
//	kit review [--base origin/main] [--labels | --json]
//	kit spec [--json]
//	kit secrets
//	kit init --owner <login> [--version <tag>]; kit init github --repo <owner/repo>; kit init app --name <name>
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
	r.AddCommand(checkCommand(out), guardCommand(out), sweepCommand(out), mutateCommand(out), reviewCommand(out), appCommand(out), hookCommand(), specCommand(out), secretsCommand(out), initCommand(out))
	return r
}

type guardOptions struct {
	base     string
	ci       bool
	approved bool
	trees    string
	asJSON   bool
}

func guardCommand(out io.Writer) *cobra.Command {
	var o guardOptions
	cmd := leaf("guard", "Check a change for bypasses, protected files and size", func(cmd *cobra.Command) error { return runGuard(cmd.Context(), out, o) })
	cmd.Flags().StringVar(&o.base, "base", "origin/main", "the branch the change goes into")
	cmd.Flags().BoolVar(&o.ci, "ci", false, "the merge check: changes that need approval fail until --approved")
	cmd.Flags().BoolVar(&o.approved, "approved", false, "a code owner approved the change's latest commit")
	cmd.Flags().StringVar(&o.trees, "approved-trees", "", "a directory of the trees of commits a code owner approved: the protected files they hold unchanged stay approved")
	cmd.Flags().BoolVar(&o.asJSON, "json", false, "print the findings as JSON")
	return cmd
}

func runGuard(ctx context.Context, out io.Writer, o guardOptions) error {
	d, findings, err := guard.Run(ctx, ".", o.base)
	if err != nil {
		return err
	}
	trees, err := guard.ReadTrees(o.trees)
	if err != nil {
		return err
	}
	covered, err := guard.Covered(ctx, findings, guard.Approval{Dir: ".", Head: o.approved, Trees: trees})
	if err != nil {
		return err
	}
	blocking := guard.Blocking(findings, o.ci, covered)
	if o.asJSON {
		err = writeJSON(out, findings)
	} else {
		err = o.report(out, findings, len(d.Files))
	}
	if err != nil {
		return err
	}
	if len(blocking) > 0 {
		return errFindings
	}
	return nil
}

func (o guardOptions) report(out io.Writer, findings []guard.Finding, files int) error {
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
		files, o.base, len(broken), len(approval))
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

// WHY: every kit command takes flags and no argument; one helper builds them.
func leaf(use, short string, run func(*cobra.Command) error) *cobra.Command {
	return &cobra.Command{Use: use, Short: short, Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error { return run(cmd) }}
}
