package domain

import "fmt"

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

// WHY: 0 is the value of an argument the caller left out, so it means the
// default rather than an empty listing.
func ParseLimit(n int) (int, error) {
	switch {
	case n == 0:
		return DefaultLimit, nil
	case n < 0 || n > MaxLimit:
		return 0, fmt.Errorf("limit %d is out of range: give 1 to %d", n, MaxLimit)
	}
	return n, nil
}
