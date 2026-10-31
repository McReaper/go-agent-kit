package add

type Args struct {
	Title string   `json:"title"`
	Tags  []string `json:"tags,omitempty"`
}

type Out struct {
	ID string `json:"id"`
}
