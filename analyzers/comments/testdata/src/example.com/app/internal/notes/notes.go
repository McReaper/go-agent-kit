// Package notes is a contract: its exported API may carry doc comments.
package notes

// Module is documented: allowed in a contract.
type Module struct {
	// Name is an exported field: allowed.
	Name string
	// count is unexported // want `M2: a comment says why`
	count int
}

// Open is documented: allowed.
func Open() *Module {
	// opens the module // want `M2: a comment says why`
	return &Module{count: 1}
}

// helper says what it does // want `M2: a comment says why`
func helper() int { return 0 }

func Use() int {
	// WHY: the zero module is valid, and callers rely on it.
	return helper()
}

//go:generate echo directives are not prose
