package m2case

// Sum adds the numbers.
func Sum(xs []int) int {
	total := 0
	for _, x := range xs {
		total += x // add x to the total
	}
	// WHY: an empty list sums to zero, which callers count on.
	return total
}
