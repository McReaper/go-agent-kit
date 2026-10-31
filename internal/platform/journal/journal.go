// Package journal keeps the app's journal: one line for every operation
// that changed something, appended, never rewritten (criterion S11). It
// answers "what did the agent do here, and when" after the fact.
package journal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

// File is the journal's name in the app's data directory.
const File = "journal.jsonl"

// Entry is one line of the journal.
type Entry struct {
	Time  time.Time       `json:"time"`
	Op    string          `json:"op"`
	Args  json.RawMessage `json:"args"`
	Error string          `json:"error,omitempty"`
}

// Journal appends entries to the journal file. It is safe for concurrent
// use: an MCP client may call tools in parallel.
type Journal struct {
	mu  sync.Mutex
	f   *os.File
	now func() time.Time
}

// Open opens the journal in dir, creating both if needed. now dates the
// entries.
func Open(dir string, now func() time.Time) (*Journal, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("create the data directory: %w", err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("open the data directory: %w", err)
	}
	f, err := root.OpenFile(File, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if err := errors.Join(err, root.Close()); err != nil {
		return nil, fmt.Errorf("open the journal: %w", err)
	}
	return &Journal{f: f, now: now}, nil
}

// Record appends that op ran with args, and failed with failure unless it
// is nil.
func (j *Journal) Record(op string, args any, failure error) error {
	a, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("journal %s: encode its arguments: %w", op, err)
	}
	e := Entry{Time: j.now().UTC(), Op: op, Args: a}
	if failure != nil {
		e.Error = failure.Error()
	}
	line, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("journal %s: %w", op, err)
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	if _, err := j.f.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("journal %s: %w", op, err)
	}
	return nil
}

// Close closes the journal file.
func (j *Journal) Close() error {
	return j.f.Close()
}
