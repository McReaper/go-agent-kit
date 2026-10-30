package m5case

import "strconv"

type Store interface {
	Get(id string) string
	Put(id, v string)
	Delete(id string)
	List() []string
	Count() int
	Close() error
}

func helper() int {
	return 1
}

func scale(x, factor int) int {
	return x * 2
}

func Doubled(a, b int) int {
	return scale(a, 2) + scale(b, 2)
}

func Parsed(s string) int {
	n, err := strconv.Atoi(s)
	n = 3
	if err != nil {
		return 0
	}
	return n
}

func Label(s string) string {
	return string(s)
}

func Last(xs []int) int {
	v := 0
	for _, x := range xs {
		v = x
	}
	v = 1
	return v
}
