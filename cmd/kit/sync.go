package main

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/McReaper/go-agent-kit/internal/platform/gomod"
	"github.com/McReaper/go-agent-kit/kitsync"
	"github.com/spf13/cobra"
)

func syncCommand(out io.Writer) *cobra.Command {
	var version, from string
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Bring the kit's files in this app to a kit version: its skills, hooks, kit.yml's pin and AGENTS.md's kit section",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runSync(cmd.Context(), out, version, from)
		},
	}
	cmd.Flags().StringVar(&version, "version", "", "the kit's commit id to sync to, all 40 hexadecimal characters")
	cmd.Flags().StringVar(&from, "from", "", "a template directory to sync from, instead of the template at --version")
	return cmd
}

func runSync(ctx context.Context, out io.Writer, version, from string) error {
	if err := kitsync.CheckVersion(version); err != nil {
		return err
	}
	if from == "" {
		dir, err := templateAt(ctx, version)
		if err != nil {
			return err
		}
		from = dir
	}
	changes, err := kitsync.Sync(".", from, version)
	if err != nil {
		return err
	}
	var b strings.Builder
	for _, c := range changes {
		fmt.Fprintf(&b, "%s %s\n", c.Action, c.Path)
	}
	files := "files"
	if len(changes) == 1 {
		files = "file"
	}
	fmt.Fprintf(&b, "kit sync: %d %s changed, kit %s; the diff is the owner's to review\n", len(changes), files, version)
	if _, err := io.WriteString(out, b.String()); err != nil {
		return fmt.Errorf("print the changes: %w", err)
	}
	return nil
}

// WHY: the template is a module of its own in the kit's repository; the
// module cache holds it at any commit, the way gonew fetched it.
func templateAt(ctx context.Context, version string) (string, error) {
	return gomod.Download(ctx, kitsync.Template, version)
}
