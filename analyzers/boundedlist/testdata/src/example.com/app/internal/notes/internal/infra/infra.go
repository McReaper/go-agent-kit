package infra

// The infra isn't an operation: a list it returns is bounded by the slice
// that asks for it.
type Out struct {
	Rows []string
}
