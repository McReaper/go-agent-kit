package m1case

func Grade(a, b, c int) string {
	if a > 0 && b > 0 {
		return "x"
	}
	if b > 1 || c > 1 {
		return "y"
	}
	if c > 2 && a > 2 {
		return "z"
	}
	if a > 3 || b > 3 {
		return "w"
	}
	if b > 4 && c > 4 {
		return "v"
	}
	if c > 5 {
		return "u"
	}
	return ""
}

func Deep(rows [][]int) int {
	n := 0
	for _, row := range rows {
		for _, x := range row {
			if x > 0 {
				n++
			}
		}
	}
	return n
}

func Five(a, b, c, d, e int) int {
	return a + b + c + d + e
}

func Long() int {
	n := 0
	n++
	n++
	n++
	n++
	n++
	n++
	n++
	n++
	n++
	n++
	n++
	n++
	n++
	n++
	n++
	n++
	n++
	n++
	n++
	n++
	n++
	n++
	n++
	n++
	n++
	n++
	n++
	n++
	n++
	return n
}
