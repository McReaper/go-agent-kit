package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/McReaper/go-agent-kit/review"
	"github.com/spf13/cobra"
)

type reviewOptions struct {
	base   string
	labels bool
	asJSON bool
}

func reviewCommand(out io.Writer) *cobra.Command {
	var o reviewOptions
	cmd := &cobra.Command{
		Use:   "review",
		Short: "Write the change's review sheet: what the owner reads instead of the diff",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runReview(cmd.Context(), out, o)
		},
	}
	cmd.Flags().StringVar(&o.base, "base", "origin/main", "the branch the change goes into")
	cmd.Flags().BoolVar(&o.labels, "labels", false, "print the kit's labels instead, +name to add, -name to remove")
	cmd.Flags().BoolVar(&o.asJSON, "json", false, "print the sheet's items as JSON")
	cmd.MarkFlagsMutuallyExclusive("labels", "json")
	return cmd
}

func runReview(ctx context.Context, out io.Writer, o reviewOptions) error {
	dir, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("review: %w", err)
	}
	s, err := review.Build(ctx, dir, o.base)
	if err != nil {
		return fmt.Errorf("review: %w", err)
	}
	switch {
	case o.labels:
		_, err = io.WriteString(out, labelLines(s.Labels()))
	case o.asJSON:
		err = json.NewEncoder(out).Encode(reviewRows(s))
	default:
		_, err = io.WriteString(out, s.Markdown())
	}
	if err != nil {
		return fmt.Errorf("print the review sheet: %w", err)
	}
	return nil
}

func labelLines(labels map[string]bool) string {
	var b strings.Builder
	for _, name := range slices.Sorted(maps.Keys(labels)) {
		sign := "-"
		if labels[name] {
			sign = "+"
		}
		b.WriteString(sign + name + "\n")
	}
	return b.String()
}

type reviewRow struct {
	Section string `json:"section"`
	File    string `json:"file"`
	Line    int    `json:"line"`
	Text    string `json:"text"`
}

func reviewRows(s review.Sheet) []reviewRow {
	rows := []reviewRow{}
	add := func(section string, items []review.Item) {
		for _, it := range items {
			rows = append(rows, reviewRow{Section: section, File: it.File, Line: it.Line, Text: it.Text})
		}
	}
	add("approval", s.Approval)
	add("test-added", s.Tests.Added)
	add("test-removed", s.Tests.Removed)
	add("test-changed", s.Tests.Changed)
	add("contract", s.Contract)
	add("api", s.API)
	for _, m := range s.Modules {
		rows = append(rows, reviewRow{Section: "module", File: m, Text: "new module"})
	}
	add("dependency", s.Deps)
	add("shared", s.Shared)
	add("rules", s.Rules)
	add("why", s.Why)
	return rows
}
