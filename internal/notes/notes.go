// Package notes keeps short notes on disk and serves them as MCP tools and
// CLI commands. It is the module's contract: nothing else in it is visible
// outside.
package notes

import (
	"errors"
	"fmt"
	"io"
	"time"

	"example.com/notes/internal/notes/internal/add"
	"example.com/notes/internal/notes/internal/edit"
	"example.com/notes/internal/notes/internal/infra"
	"example.com/notes/internal/notes/internal/list"
	"example.com/notes/internal/platform/op"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

// Module is the notes module, ready to be registered on both surfaces.
type Module struct {
	store *infra.FileStore
	add   *add.Add
	list  *list.List
	edit  *edit.Edit
}

// Open opens the notes kept in dir, creating it if needed. now is the clock
// new notes are dated with.
func Open(dir string, now func() time.Time) (*Module, error) {
	store, err := infra.OpenFileStore(dir)
	if err != nil {
		return nil, err
	}
	return &Module{store: store, add: add.New(store, now), list: list.New(store), edit: edit.New(store)}, nil
}

// Close releases the notes directory.
func (m *Module) Close() error {
	return m.store.Close()
}

// Register adds the module's tools to s.
func (m *Module) Register(s *mcp.Server) error {
	return errors.Join(
		op.Register(s, m.add.Op()),
		op.Register(s, m.list.Op()),
		op.Register(s, m.edit.Op()),
	)
}

// Commands returns the module's CLI commands, which print their results to out.
func (m *Module) Commands(out io.Writer) ([]*cobra.Command, error) {
	addCmd, aerr := op.Command(m.add.Op(), out)
	listCmd, lerr := op.Command(m.list.Op(), out)
	editCmd, eerr := op.Command(m.edit.Op(), out)
	if err := errors.Join(aerr, lerr, eerr); err != nil {
		return nil, fmt.Errorf("notes commands: %w", err)
	}
	return []*cobra.Command{addCmd, listCmd, editCmd}, nil
}
