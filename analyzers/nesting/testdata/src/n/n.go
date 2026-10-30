package n

func Ok(rows [][]int) int {
	n := 0
	for _, r := range rows {
		if len(r) > 0 {
			n++
		} else if len(r) == 0 {
			n--
		}
	}
	return n
}

func Deep(rows [][]int) int {
	n := 0
	for _, r := range rows {
		for _, x := range r {
			if x > 0 { // want `M1: control flow nested 3 deep, the limit is 2`
				n++
			}
		}
	}
	return n
}

func Switch(xs []int) int {
	n := 0
	for _, x := range xs {
		switch {
		case x > 0:
			if x > 9 { // want `M1: control flow nested 3 deep`
				n++
			}
		}
	}
	return n
}

func Closure(xs []int) func() {
	return func() {
		for range xs {
			if len(xs) > 1 {
				select {} // want `M1: control flow nested 3 deep`
			}
		}
	}
}
