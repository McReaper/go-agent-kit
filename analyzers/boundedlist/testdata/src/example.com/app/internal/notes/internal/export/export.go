package export

type Args struct {
	Max   int    `json:"max"`
	Limit string `json:"limit"`
}

type Out struct {
	Rows map[string]string `json:"rows"` // want `S8: Out.Rows is a list, and Args has no limit`
}
