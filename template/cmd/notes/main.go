package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/McReaper/go-agent-kit/template/internal/notes"
	"github.com/McReaper/go-agent-kit/template/internal/platform/paths"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	err := run(ctx, os.Stdout)
	stop()
	if err != nil {
		log.ErrorContext(ctx, "notes failed", slog.Any("err", err))
		os.Exit(1)
	}
}

func run(ctx context.Context, out io.Writer) (err error) {
	dir, err := paths.DataDir("notes")
	if err != nil {
		return err
	}
	m, err := notes.Open(dir, time.Now)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, m.Close()) }()
	root, err := rootCommand(m, out)
	if err != nil {
		return err
	}
	return root.ExecuteContext(ctx)
}

func rootCommand(m *notes.Module, out io.Writer) (*cobra.Command, error) {
	root := &cobra.Command{Use: "notes", Short: "Short notes for people and agents", Version: version, SilenceUsage: true}
	cmds, err := m.Commands(out)
	if err != nil {
		return nil, err
	}
	root.AddCommand(cmds...)
	root.AddCommand(&cobra.Command{
		Use:   "mcp",
		Short: "Serve the notes as MCP tools on stdin and stdout",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			s := mcp.NewServer(&mcp.Implementation{Name: "notes", Version: version}, nil)
			if err := m.Register(s); err != nil {
				return err
			}
			return s.Run(cmd.Context(), &mcp.StdioTransport{})
		},
	})
	return root, nil
}
