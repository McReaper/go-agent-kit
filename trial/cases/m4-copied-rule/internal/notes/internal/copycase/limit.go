package copycase

import "fmt"

const maxLimit = 100

func LimitOfList(p *int) (int, error) {
	limit := 20
	if p != nil {
		limit = *p
	}
	if limit < 1 || limit > maxLimit {
		return 0, fmt.Errorf("limit %d is out of range: give 1 to %d", limit, maxLimit)
	}
	return limit, nil
}

func LimitOfSearch(p *int) (int, error) {
	limit := 20
	if p != nil {
		limit = *p
	}
	if limit < 1 || limit > maxLimit {
		return 0, fmt.Errorf("limit %d is out of range: give 1 to %d", limit, maxLimit)
	}
	return limit, nil
}
