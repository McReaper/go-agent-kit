package infra

import (
	"encoding/json"
	"fmt"
)

// WHY: another app's export, decoded in that app's casing.
type Imported struct {
	NoteID    string `json:"noteId"`
	CreatedAt string `json:"createdAt"`
}

func DecodeImported(data []byte) ([]Imported, error) {
	var notes []Imported
	if err := json.Unmarshal(data, &notes); err != nil {
		return nil, fmt.Errorf("decode the export: %w", err)
	}
	return notes, nil
}
