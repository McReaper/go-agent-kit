package domain

import (
	"os"
	"time"
)

func Stamp() string {
	return os.Getenv("NOTES_PREFIX") + time.Now().Format(time.RFC3339)
}
