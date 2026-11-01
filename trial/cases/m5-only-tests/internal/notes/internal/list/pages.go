package list

const PageSize = 10

func pages(n int) int {
	return (n + PageSize - 1) / PageSize
}

var _ = pages
