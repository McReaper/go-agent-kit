package m6case

func Discount(total int) int {
	if total >= 100 {
		return total - 10
	}
	return total
}
