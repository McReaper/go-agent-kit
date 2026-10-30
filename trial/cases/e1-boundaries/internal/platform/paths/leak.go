package paths

import "example.com/notes/internal/notes"

func Leak(m *notes.Module) bool {
	return m != nil
}
