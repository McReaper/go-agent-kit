package paths

import "github.com/McReaper/go-agent-kit/template/internal/notes"

func Leak(m *notes.Module) bool {
	return m != nil
}
