package domain

import (
	"fmt"
	"strings"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

// WHY: the limit rule lives here so every operation that returns a list of
// notes answers a bad limit the same way.
func CheckLimit(limit int) (int, error) {
	switch {
	case limit == 0:
		return DefaultLimit, nil
	case limit < 0 || limit > MaxLimit:
		return 0, fmt.Errorf("limit %d is out of range: give 1 to %d", limit, MaxLimit)
	}
	return limit, nil
}

// WHY: both sides are lowered, so "MILK" finds a title that says "milk".
func (n Note) Contains(text string) bool {
	needle := strings.ToLower(text)
	return strings.Contains(strings.ToLower(string(n.Title)), needle) ||
		strings.Contains(strings.ToLower(n.Body), needle)
}
