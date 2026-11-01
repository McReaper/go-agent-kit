package list

type Args struct {
	Limit *int   `json:"limit,omitempty"`
	Tag   string `json:"tag,omitempty"`
}

type Out struct {
	Notes []string `json:"notes"`
}
