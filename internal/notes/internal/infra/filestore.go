package infra

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"
	"time"

	"example.com/notes/internal/notes/internal/domain"
)

const ext = ".json"

type record struct {
	ID      string    `json:"id"`
	Title   string    `json:"title"`
	Body    string    `json:"body"`
	Created time.Time `json:"created"`
}

type FileStore struct {
	root *os.Root
}

func OpenFileStore(dir string) (*FileStore, error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("create the notes directory: %w", err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("open the notes directory: %w", err)
	}
	return &FileStore{root: root}, nil
}

func (s *FileStore) Close() error {
	return s.root.Close()
}

func (s *FileStore) Save(_ context.Context, n domain.Note) error {
	data, err := encode(n)
	if err != nil {
		return err
	}
	f, err := s.root.OpenFile(string(n.ID)+ext, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("note %s already exists", n.ID)
	}
	if err != nil {
		return fmt.Errorf("create note %s: %w", n.ID, err)
	}
	_, werr := f.Write(data)
	cerr := f.Close()
	if err := errors.Join(werr, cerr); err != nil {
		return fmt.Errorf("write note %s: %w", n.ID, err)
	}
	return nil
}

// WHY: a missing note is reported as domain.ErrNoteNotFound, so callers
// tell "no such id" from a broken store without reading file errors.
func (s *FileStore) Get(_ context.Context, id domain.NoteID) (domain.Note, error) {
	n, err := s.read(string(id) + ext)
	if errors.Is(err, fs.ErrNotExist) {
		return domain.Note{}, fmt.Errorf("%w: %s", domain.ErrNoteNotFound, id)
	}
	return n, err
}

// WHY: the write goes to a temporary file renamed over the note, so a crash
// leaves the old note or the new one, never half of each.
func (s *FileStore) Replace(_ context.Context, n domain.Note) error {
	name := string(n.ID) + ext
	if _, err := s.root.Stat(name); err != nil {
		return fmt.Errorf("replace note %s: %w", n.ID, err)
	}
	data, err := encode(n)
	if err != nil {
		return err
	}
	tmp := name + ".tmp"
	f, err := s.root.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("write note %s: %w", n.ID, err)
	}
	_, werr := f.Write(data)
	cerr := f.Close()
	if err := errors.Join(werr, cerr); err != nil {
		return errors.Join(fmt.Errorf("write note %s: %w", n.ID, err), s.root.Remove(tmp))
	}
	if err := s.root.Rename(tmp, name); err != nil {
		return fmt.Errorf("replace note %s: %w", n.ID, err)
	}
	return nil
}

func encode(n domain.Note) ([]byte, error) {
	data, err := json.Marshal(record{ID: string(n.ID), Title: string(n.Title), Body: n.Body, Created: n.Created})
	if err != nil {
		return nil, fmt.Errorf("encode note %s: %w", n.ID, err)
	}
	return data, nil
}

// WHY: All sorts the notes, so Newest and search both get newest first from
// the one order.
func (s *FileStore) All(_ context.Context) ([]domain.Note, error) {
	names, err := fs.Glob(s.root.FS(), "*"+ext)
	if err != nil {
		return nil, fmt.Errorf("list the notes: %w", err)
	}
	notes := make([]domain.Note, 0, len(names))
	for _, name := range names {
		n, err := s.read(name)
		if err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	slices.SortFunc(notes, func(a, b domain.Note) int {
		if c := b.Created.Compare(a.Created); c != 0 {
			return c
		}
		return strings.Compare(string(a.ID), string(b.ID))
	})
	return notes, nil
}

func (s *FileStore) Newest(ctx context.Context, limit int) ([]domain.Note, error) {
	notes, err := s.All(ctx)
	if err != nil {
		return nil, err
	}
	return notes[:min(limit, len(notes))], nil
}

func (s *FileStore) read(name string) (domain.Note, error) {
	data, err := s.root.ReadFile(name)
	if err != nil {
		return domain.Note{}, fmt.Errorf("read %s: %w", name, err)
	}
	var r record
	if err := json.Unmarshal(data, &r); err != nil {
		return domain.Note{}, fmt.Errorf("decode %s: %w", name, err)
	}
	return domain.Note{ID: domain.NoteID(r.ID), Title: domain.Title(r.Title), Body: r.Body, Created: r.Created}, nil
}
