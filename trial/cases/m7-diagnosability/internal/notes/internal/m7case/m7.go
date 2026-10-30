package m7case

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
)

var ErrGone = errors.New("gone")

func Read(r io.Reader) error {
	_, err := r.Read(make([]byte, 1))
	if err == ErrGone {
		return nil
	}
	slog.Info("read", "bytes", 1)
	return fmt.Errorf("read: %v", err)
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Quit() {
	os.Exit(2)
}

func Guard(f func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered: %v", r)
		}
	}()
	f()
	return nil
}
