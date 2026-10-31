package infra

import (
	"context"

	"example.com/notes/internal/notes/internal/domain"
)

func (s *FileStore) Oldest(ctx context.Context) (domain.Note, error) {
	notes, err := s.Newest(ctx, 100)
	if err != nil || len(notes) == 0 {
		return domain.Note{}, err
	}
	return notes[len(notes)-1], nil
}
