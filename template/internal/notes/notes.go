// Package notes keeps short notes on disk and serves them as MCP tools and
// CLI commands. It is the module's contract: nothing else in it is visible
// outside.
package notes

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/add"
	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/infra"
	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/list"
	"github.com/McReaper/go-agent-kit/template/internal/platform/op"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

// Module is the notes module, ready to be registered on both surfaces.
type Module struct {
	store *infra.FileStore
	add   *add.Add
	list  *list.List
}

// Open opens the notes kept in dir, creating it if needed. now is the clock
// new notes are dated with.
func Open(dir string, now func() time.Time) (*Module, error) {
	store, err := infra.OpenFileStore(dir)
	if err != nil {
		return nil, err
	}
	return &Module{store: store, add: add.New(store, now), list: list.New(store)}, nil
}

// Close releases the notes directory.
func (m *Module) Close() error {
	return m.store.Close()
}

// Register adds the module's tools to s, journaling in j those that change
// something.
func (m *Module) Register(s *mcp.Server, j op.Journal) error {
	return errors.Join(op.Register(s, j, m.add.Op()), op.Register(s, j, m.list.Op()))
}

// Commands returns the module's CLI commands, which journal in j those that
// change something and print their results to out.
func (m *Module) Commands(j op.Journal, out io.Writer) ([]*cobra.Command, error) {
	addCmd, aerr := op.Command(m.add.Op(), j, out)
	listCmd, lerr := op.Command(m.list.Op(), j, out)
	if err := errors.Join(aerr, lerr); err != nil {
		return nil, fmt.Errorf("notes commands: %w", err)
	}
	return []*cobra.Command{addCmd, listCmd}, nil
}
