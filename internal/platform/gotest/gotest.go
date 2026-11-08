// Package gotest reads what go test -json writes.
package gotest

import (
	"bufio"
	"bytes"
	"encoding/json"
)

// Event is one line of go test -json: an action on a package, or on one of
// its tests, and the output that came with it.
type Event struct {
	Action  string `json:"Action"`
	Package string `json:"Package"`
	Test    string `json:"Test"`
	Output  string `json:"Output"`
}

// Events decodes stream, skipping any line that isn't an event.
//
// WHY: go test prints a build failure's own lines outside the JSON in some
// versions; they are the compiler's, and the events say the same.
func Events(stream []byte) []Event {
	var out []Event
	sc := bufio.NewScanner(bytes.NewReader(stream))
	sc.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for sc.Scan() {
		var e Event
		if json.Unmarshal(sc.Bytes(), &e) == nil {
			out = append(out, e)
		}
	}
	return out
}
