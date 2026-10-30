// Package add is inside a module: no doc comments here. // want `M2: a comment says why`
package add

// Run is exported but not a contract. // want `M2: a comment says why`
func Run() int {
	/* WHY: a block comment may carry a reason too. */
	return 1 // returns one // want `M2: a comment says why`
}
