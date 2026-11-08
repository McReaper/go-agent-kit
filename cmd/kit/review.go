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

	"github.com/McReaper/go-agent-kit/guard"
	"github.com/McReaper/go-agent-kit/review"
	"github.com/spf13/cobra"
)

type reviewOptions struct {
	base     string
	labels   bool
	asJSON   bool
	approval bool
	approved bool
	trees    string
}

func reviewCommand(out io.Writer) *cobra.Command {
	var o reviewOptions
	cmd := leaf("review", "Write the change's review sheet: what the owner reads instead of the diff", func(cmd *cobra.Command) error { return runReview(cmd.Context(), out, o) })
	cmd.Flags().StringVar(&o.base, "base", "origin/main", "the branch the change goes into")
	cmd.Flags().BoolVar(&o.labels, "labels", false, "print the kit's labels instead, +name to add, -name to remove")
	cmd.Flags().BoolVar(&o.asJSON, "json", false, "print the sheet's items as JSON")
	cmd.Flags().BoolVar(&o.approval, "approval", false, "print where the owner's approval stands instead: none, waiting or approved")
	cmd.Flags().BoolVar(&o.approved, "approved", false, "a code owner approved the change's latest commit")
	cmd.Flags().StringVar(&o.trees, "approved-trees", "", "a directory of the trees of commits a code owner approved: the protected files they hold unchanged stay approved")
	cmd.MarkFlagsMutuallyExclusive("labels", "json", "approval")
	return cmd
}

func runReview(ctx context.Context, out io.Writer, o reviewOptions) error {
	dir, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf("review: %w", err)
	}
	trees, err := guard.ReadTrees(o.trees)
	if err != nil {
		return fmt.Errorf("review: %w", err)
	}
	s, err := review.Build(ctx, dir, o.base)
	if err != nil {
		return fmt.Errorf("review: %w", err)
	}
	if s, err = s.Approve(ctx, guard.Approval{Dir: dir, Head: o.approved, Trees: trees}); err != nil {
		return fmt.Errorf("review: %w", err)
	}
	switch {
	case o.labels:
		_, err = io.WriteString(out, labelLines(s.Labels()))
	case o.asJSON:
		err = json.NewEncoder(out).Encode(reviewRows(s))
	case o.approval:
		_, err = io.WriteString(out, string(s.State)+"\n")
	default:
		_, err = io.WriteString(out, s.Markdown(ownVersion()))
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
	for _, c := range s.Spec {
		rows = append(rows, reviewRow{Section: "spec-" + string(c.Kind), File: c.File, Line: c.Line, Text: c.Name})
	}
	add("approval", s.Approval)
	return rows
}
