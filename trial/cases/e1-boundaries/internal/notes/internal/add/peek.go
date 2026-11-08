package add

import (
	"github.com/McReaper/rumblestrip-go/template/internal/notes/internal/infra"
	"github.com/McReaper/rumblestrip-go/template/internal/notes/internal/list"
)

func Peek(s *infra.FileStore) *list.List {
	return list.New(s)
}
