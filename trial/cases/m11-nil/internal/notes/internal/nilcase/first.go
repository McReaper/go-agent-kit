package nilcase

type Item struct {
	Name string
}

func First(items []*Item) *Item {
	if len(items) == 0 {
		return nil
	}
	return items[0]
}

func FirstName(items []*Item) string {
	return First(items).Name
}
