package domain

import "time"

func (n Note) Age() time.Duration {
	return time.Since(n.Created)
}
