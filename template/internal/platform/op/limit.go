package op

import "fmt"

// Limit is the limit an operation that returns a list was given: def when
// it wasn't, an error naming the range when it is outside 1 to max.
//
// WHY: every operation that returns a list takes a limit (kitvet
// boundedlist); the rule has one home, here, so no operation copies it.
func Limit(given *int, def, maxLimit int) (int, error) {
	limit := def
	if given != nil {
		limit = *given
	}
	if limit < 1 || limit > maxLimit {
		return 0, fmt.Errorf("limit %d is out of range: give 1 to %d", limit, maxLimit)
	}
	return limit, nil
}
