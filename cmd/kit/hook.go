package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/McReaper/go-agent-kit/check"
	"github.com/spf13/cobra"
)

// WHY: the process exits with this code, not 1: Claude Code reads a
// hook's code 2 as "show my stderr to Claude", or for Stop "don't stop".
type exitCode int

func (c exitCode) Error() string { return fmt.Sprintf("exit %d", int(c)) }

type hookInput struct {
	Cwd            string `json:"cwd"`
	StopHookActive bool   `json:"stop_hook_active"`
	ToolInput      struct {
		FilePath string `json:"file_path"`
	} `json:"tool_input"`
}

func readHookInput(in io.Reader) (hookInput, error) {
	var h hookInput
	if err := json.NewDecoder(in).Decode(&h); err != nil {
		return h, fmt.Errorf("read the hook's input: %w", err)
	}
	return h, nil
}

func hookCommand() *cobra.Command {
	hook := &cobra.Command{Use: "hook", Short: "Claude Code hooks: feedback after each edit, the full check before stopping"}
	var base string
	stop := &cobra.Command{
		Use:   "stop",
		Short: "Run kit check before the agent stops; a failing gate keeps it working, once",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return stopHook(cmd, base)
		},
	}
	stop.Flags().StringVar(&base, "base", "origin/main", "the branch the change goes into")
	hook.AddCommand(&cobra.Command{
		Use:   "post-edit",
		Short: "Lint the package of the Go file just edited; findings go back to the agent",
		Args:  cobra.NoArgs,
		RunE:  postEditHook,
	}, stop)
	return hook
}

func postEditHook(cmd *cobra.Command, _ []string) error {
	h, err := readHookInput(cmd.InOrStdin())
	if err != nil || !strings.HasSuffix(h.ToolInput.FilePath, ".go") {
		return err
	}
	root, err := moduleRoot(filepath.Dir(h.ToolInput.FilePath))
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(root, filepath.Dir(h.ToolInput.FilePath))
	if err != nil {
		return fmt.Errorf("post-edit: %w", err)
	}
	o := check.Options{Dir: root, GOOS: []string{"linux", "windows"}}
	return feedback(cmd.ErrOrStderr(), check.Package(cmd.Context(), o, "./"+filepath.ToSlash(rel)))
}

func stopHook(cmd *cobra.Command, base string) error {
	h, err := readHookInput(cmd.InOrStdin())
	if err != nil || h.StopHookActive {
		return err
	}
	o := check.Options{Dir: h.Cwd, Base: base, GOOS: []string{"linux", "windows"}}
	if root, err := moduleRoot(h.Cwd); err == nil {
		o.Dir = root
	}
	return feedback(cmd.ErrOrStderr(), check.Run(cmd.Context(), o))
}

// WHY: only failures are written, to stderr with code 2, the one channel
// Claude Code shows the agent; a clean run says nothing.
func feedback(stderr io.Writer, results []check.Result) error {
	var b strings.Builder
	failing := 0
	for _, r := range results {
		if r.Failed() {
			failing++
			writeResult(&b, r)
		}
	}
	if failing == 0 {
		return nil
	}
	fmt.Fprintf(&b, "kit: %d gates fail; fix them before you finish. Each line says the rule and how to fix it.\n", failing)
	if _, err := io.WriteString(stderr, b.String()); err != nil {
		return fmt.Errorf("write the hook's feedback: %w", err)
	}
	return exitCode(2)
}

func moduleRoot(dir string) (string, error) {
	for d := dir; ; d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, "go.mod")); err == nil {
			return d, nil
		}
		if filepath.Dir(d) == d {
			return "", fmt.Errorf("no go.mod above %s", dir)
		}
	}
}
