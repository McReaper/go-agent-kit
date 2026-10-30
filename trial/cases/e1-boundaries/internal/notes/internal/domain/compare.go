package domain

import "github.com/google/go-cmp/cmp"

func Same(a, b Note) bool {
	return cmp.Equal(a, b)
}
