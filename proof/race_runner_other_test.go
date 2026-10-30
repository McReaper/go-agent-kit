//go:build !linux

package proof_test

// WHY: the race detector needs cgo; off Linux its cases are proven by CI only.
const raceRunner = false
