package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/McReaper/go-agent-kit/spec"
	"github.com/spf13/cobra"
)

func specCommand(out io.Writer) *cobra.Command {
	var asJSON bool
	cmd := leaf("spec", "Prove the spec: each scenario passes alone, and fails when a value a Then step expects is altered", func(cmd *cobra.Command) error { return runSpec(cmd.Context(), out, asJSON) })
	cmd.Flags().BoolVar(&asJSON, "json", false, "print the findings as JSON")
	return cmd
}

func runSpec(ctx context.Context, out io.Writer, asJSON bool) error {
	dir, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("spec: %w", err)
	}
	fs, err := spec.Prove(ctx, dir)
	if err != nil {
		return fmt.Errorf("spec: %w", err)
	}
	if asJSON {
		err = writeSpecJSON(out, fs)
	} else {
		_, err = io.WriteString(out, specReport(fs))
	}
	if err != nil {
		return fmt.Errorf("print the spec's findings: %w", err)
	}
	if len(fs) > 0 {
		return errFindings
	}
	return nil
}

func specReport(fs []spec.Finding) string {
	var b strings.Builder
	for _, l := range spec.Lines(fs) {
		b.WriteString(l + "\n")
	}
	fmt.Fprintf(&b, "kit spec: %d findings\n", len(fs))
	return b.String()
}

func writeSpecJSON(out io.Writer, fs []spec.Finding) error {
	type row struct {
		Rule    string `json:"rule"`
		File    string `json:"file"`
		Line    int    `json:"line"`
		Message string `json:"message"`
	}
	rows := make([]row, 0, len(fs))
	for _, f := range fs {
		rows = append(rows, row(f))
	}
	return json.NewEncoder(out).Encode(rows)
}
