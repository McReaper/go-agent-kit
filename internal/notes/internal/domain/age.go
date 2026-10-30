package domain

import "time"

func (n Note) Age(now time.Time) time.Duration {
	return now.Sub(n.Created)
}
