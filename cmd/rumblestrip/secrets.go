package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/McReaper/rumblestrip-go/internal/platform/gitleaks"
	"github.com/spf13/cobra"
)

// WHY: a secret stopped before the commit never enters the history, which
// only a rewrite cleans; rumblestrip check scans the commits too, in CI.
func secretsCommand(out io.Writer) *cobra.Command {
	return leaf("secrets", "Scan the staged lines for secrets, keys and env files, with Rumblestrip's gitleaks rules", func(cmd *cobra.Command) error {
		leaks, err := gitleaks.Scan(cmd.Context(), ".", "--pre-commit", "--staged")
		if err != nil {
			return err
		}
		var b strings.Builder
		for _, l := range leaks {
			fmt.Fprintf(&b, "%s:%d: %s: %s\n", l.File, l.Line, l.Rule, l.Description)
		}
		fmt.Fprintf(&b, "rumblestrip secrets: %d found in the staged lines\n", len(leaks))
		if _, err := io.WriteString(out, b.String()); err != nil {
			return fmt.Errorf("print the secrets found: %w", err)
		}
		if len(leaks) > 0 {
			return errFindings
		}
		return nil
	})
}
