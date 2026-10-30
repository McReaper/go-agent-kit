package n

import "testing"

func TestDeepIsExempt(t *testing.T) {
	for range 1 {
		for range 1 {
			if t != nil {
				return
			}
		}
	}
}
