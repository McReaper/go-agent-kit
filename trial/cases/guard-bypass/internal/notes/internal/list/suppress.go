package list

import _ "unsafe"

//go:linkname nanotime runtime.nanotime
func nanotime() int64

func Quiet() int { //nolint:gocognit
	return int(nanotime()) // #nosec G115
}
