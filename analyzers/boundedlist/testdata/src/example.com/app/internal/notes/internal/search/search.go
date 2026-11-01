package search

type Args struct {
	Text string `json:"text"`
}

type Item struct {
	ID string `json:"id"`
}

type Out struct {
	Count int    `json:"count"`
	Notes []Item `json:"notes"` // want `S8: Out.Notes is a list, and Args has no limit`
}
