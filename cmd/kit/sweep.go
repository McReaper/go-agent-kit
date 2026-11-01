package main

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/McReaper/go-agent-kit/sweep"
	"github.com/spf13/cobra"
)

func sweepCommand(out io.Writer) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "sweep",
		Short: "Check the whole module: code only tests reach, over-exported names, duplicate log messages",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runSweep(out, asJSON)
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the findings as JSON")
	return cmd
}

func runSweep(out io.Writer, asJSON bool) error {
	dir, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("sweep: %w", err)
	}
	findings, err := sweep.Check(dir)
	if err != nil {
		return fmt.Errorf("sweep: %w", err)
	}
	if asJSON {
		err = writeSweepJSON(out, dir, findings)
	} else {
		err = writeSweepText(out, dir, findings)
	}
	if err != nil {
		return err
	}
	if len(findings) > 0 {
		return errFindings
	}
	return nil
}

func relative(dir, file string) string {
	if rel, err := filepath.Rel(dir, file); err == nil {
		return filepath.ToSlash(rel)
	}
	return file
}

func writeSweepText(out io.Writer, dir string, findings []sweep.Finding) error {
	var b strings.Builder
	for _, f := range findings {
		fmt.Fprintf(&b, "%s:%d: %s: %s\n", relative(dir, f.File), f.Line, f.Rule, f.Message)
	}
	if _, err := io.WriteString(out, b.String()); err != nil {
		return fmt.Errorf("print findings: %w", err)
	}
	return nil
}

func writeSweepJSON(out io.Writer, dir string, findings []sweep.Finding) error {
	type row struct {
		Rule    string `json:"rule"`
		File    string `json:"file"`
		Line    int    `json:"line"`
		Message string `json:"message"`
	}
	rows := make([]row, 0, len(findings))
	for _, f := range findings {
		rows = append(rows, row{Rule: f.Rule, File: relative(dir, f.File), Line: f.Line, Message: f.Message})
	}
	if err := json.NewEncoder(out).Encode(rows); err != nil {
		return fmt.Errorf("print findings: %w", err)
	}
	return nil
}
