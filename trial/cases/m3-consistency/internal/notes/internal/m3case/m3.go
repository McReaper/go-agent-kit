package m3case

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
)

type Ref struct {
	NoteID string `json:"noteId"`
}

func Report(r Ref) {
	log.Printf("note %s", r.NoteID)
	fmt.Println(r.NoteID)
}

func Smaller(a, b int) int {
	m := a
	if b < a {
		m = b
	}
	return m
}

func Sink() io.Writer {
	return ioutil.Discard
}
