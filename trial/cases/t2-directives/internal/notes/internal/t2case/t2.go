package t2case

import (
	"unsafe"
	_ "unsafe"
)

// go:generate echo hidden

//go:linkname runtimeNano runtime.nanotime
func runtimeNano() int64

func Size() uintptr {
	return unsafe.Sizeof(runtimeNano())
}
