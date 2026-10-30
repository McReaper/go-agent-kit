// Command kit runs the kit's checks on a repository.
//
//	kit guard --base origin/main [--approved] [--json]
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/McReaper/go-agent-kit/guard"
	"github.com/spf13/cobra"
)

var errFindings = errors.New("the change breaks the kit's rules")

func main() {
	if err := root(os.Stdout).ExecuteContext(context.Background()); err != nil {
		os.Exit(1)
	}
}

func root(out io.Writer) *cobra.Command {
	r := &cobra.Command{Use: "kit", Short: "The anti-drift kit's checks", SilenceUsage: true}
	r.AddCommand(guardCommand(out))
	return r
}

func guardCommand(out io.Writer) *cobra.Command {
	var (
		base             string
		approved, asJSON bool
	)
	cmd := &cobra.Command{
		Use:   "guard",
		Short: "Check a change for bypasses, protected files and size",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			d, err := guard.Read(cmd.Context(), ".", base)
			if err != nil {
				return err
			}
			findings := guard.Check(d, approved)
			if err := report(out, findings, asJSON); err != nil {
				return err
			}
			if len(findings) > 0 {
				return errFindings
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&base, "base", "origin/main", "the branch the change goes into")
	cmd.Flags().BoolVar(&approved, "approved", false, "a code owner approved the change's latest commit: protected files may change")
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the findings as JSON")
	return cmd
}

func report(out io.Writer, findings []guard.Finding, asJSON bool) error {
	if asJSON {
		return writeJSON(out, findings)
	}
	for _, f := range findings {
		if _, err := fmt.Fprintf(out, "%s:%d: %s: %s\n", f.File, f.Line, f.Rule, f.Message); err != nil {
			return fmt.Errorf("print findings: %w", err)
		}
	}
	return nil
}

func writeJSON(out io.Writer, findings []guard.Finding) error {
	type row struct {
		Rule    string `json:"rule"`
		File    string `json:"file"`
		Line    int    `json:"line"`
		Message string `json:"message"`
	}
	rows := make([]row, 0, len(findings))
	for _, f := range findings {
		rows = append(rows, row(f))
	}
	if err := json.NewEncoder(out).Encode(rows); err != nil {
		return fmt.Errorf("print findings: %w", err)
	}
	return nil
}
