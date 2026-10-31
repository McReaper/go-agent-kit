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
	data, err := json.Marshal(record{ID: string(n.ID), Title: string(n.Title), Body: n.Body, Created: n.Created})
	if err != nil {
		return fmt.Errorf("encode note %s: %w", n.ID, err)
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

func (s *FileStore) Newest(_ context.Context, limit int) ([]domain.Note, error) {
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
	id, err := domain.ParseNoteID(r.ID)
	if err != nil {
		return domain.Note{}, fmt.Errorf("decode %s: %w", name, err)
	}
	title, err := domain.ParseTitle(r.Title)
	if err != nil {
		return domain.Note{}, fmt.Errorf("decode %s: %w", name, err)
	}
	return domain.Note{ID: id, Title: title, Body: r.Body, Created: r.Created}, nil
}
