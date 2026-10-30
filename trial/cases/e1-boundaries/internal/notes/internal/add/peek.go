package add

import (
	"example.com/notes/internal/notes/internal/infra"
	"example.com/notes/internal/notes/internal/list"
)

func Peek(s *infra.FileStore) *list.List {
	return list.New(s)
}
