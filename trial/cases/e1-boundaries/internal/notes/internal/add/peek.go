package add

import (
	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/infra"
	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/list"
)

func Peek(s *infra.FileStore) *list.List {
	return list.New(s)
}
