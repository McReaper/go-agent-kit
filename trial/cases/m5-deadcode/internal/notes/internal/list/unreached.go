package list

func Oldest(o Out) Item {
	return o.Notes[len(o.Notes)-1]
}
