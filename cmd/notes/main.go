package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/McReaper/go-agent-kit/template/internal/notes"
	"github.com/McReaper/go-agent-kit/template/internal/platform/config"
	"github.com/McReaper/go-agent-kit/template/internal/platform/journal"
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
	cfg, err := config.Load("notes", os.Getenv)
	if err != nil {
		return err
	}
	j, err := journal.Open(cfg.DataDir, time.Now)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, j.Close()) }()
	m, err := notes.Open(filepath.Join(cfg.DataDir, "notes"), time.Now)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, m.Close()) }()
	root, err := rootCommand(m, j, out)
	if err != nil {
		return err
	}
	return root.ExecuteContext(ctx)
}

func rootCommand(m *notes.Module, j *journal.Journal, out io.Writer) (*cobra.Command, error) {
	root := &cobra.Command{Use: "notes", Short: "Short notes for people and agents", Version: version, SilenceUsage: true}
	cmds, err := m.Commands(j, out)
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
			if err := m.Register(s, j); err != nil {
				return err
			}
			return s.Run(cmd.Context(), &mcp.StdioTransport{})
		},
	})
	return root, nil
}
