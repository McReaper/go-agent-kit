package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/McReaper/rumblestrip-go/internal/platform/ghapi"
	"github.com/McReaper/rumblestrip-go/internal/platform/git"
	"github.com/McReaper/rumblestrip-go/pr"
	"github.com/spf13/cobra"
)

type prOptions struct {
	base, title, body, bodyFile string
}

func prCommand(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "pr", Short: "Pull requests the way Rumblestrip's check wants them"}
	var o prOptions
	create := leaf("create", "Open the current branch's pull request: into the default branch, or stacked on the pull request of --base", func(cmd *cobra.Command) error {
		return runPRCreate(cmd, out, o)
	})
	create.Flags().StringVar(&o.base, "base", "", "the branch it goes into (default: the default branch); another branch's pull request is the one it stacks on")
	create.Flags().StringVar(&o.title, "title", "", "the title: a conventional commit of 72 characters at most, the commit's subject once merged")
	create.Flags().StringVar(&o.body, "body", "", "the description: why, in a few lines")
	create.Flags().StringVar(&o.bodyFile, "body-file", "", "read the description from this file")
	create.MarkFlagsMutuallyExclusive("body", "body-file")
	stack := leaf("stack", "Stack the current branch's open pull request on the pull request of the branch it goes into", func(cmd *cobra.Command) error {
		return runPRStack(cmd, out)
	})
	cmd.AddCommand(create, stack)
	return cmd
}

func runPRStack(cmd *cobra.Command, out io.Writer) error {
	head, err := currentBranch(cmd)
	if err != nil {
		return err
	}
	stacked, err := pr.Stack(cmd.Context(), ghapi.GH{}, head)
	if err != nil {
		return err
	}
	return say(out, fmt.Sprintf("stacked #%d on #%d: GitHub moves it onto the default branch once #%d merges\n", stacked.Number, stacked.StackedOn, stacked.StackedOn))
}

func say(out io.Writer, text string) error {
	if _, err := io.WriteString(out, text); err != nil {
		return fmt.Errorf("pr: %w", err)
	}
	return nil
}

func currentBranch(cmd *cobra.Command) (string, error) {
	head, err := git.Run(cmd.Context(), ".", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("pr: the current branch: %w", err)
	}
	return strings.TrimSpace(head), nil
}

func runPRCreate(cmd *cobra.Command, out io.Writer, o prOptions) error {
	head, err := currentBranch(cmd)
	if err != nil {
		return err
	}
	body := o.body
	if o.bodyFile != "" {
		data, err := os.ReadFile(o.bodyFile)
		if err != nil {
			return fmt.Errorf("pr: %w", err)
		}
		body = string(data)
	}
	opened, err := pr.Open(cmd.Context(), ghapi.GH{}, pr.Options{Head: head, Base: o.base, Title: o.title, Body: body})
	var b strings.Builder
	if opened.Number != 0 {
		fmt.Fprintf(&b, "opened #%d %s\n", opened.Number, opened.URL)
	}
	if opened.StackedOn != 0 {
		fmt.Fprintf(&b, "stacked on #%d: GitHub moves it onto the default branch once #%d merges\n", opened.StackedOn, opened.StackedOn)
	}
	return errors.Join(say(out, b.String()), err)
}
