package infra_test

import (
	"context"
	"testing"
)

func TestOldestIsTheFirstSaved(t *testing.T) {
	s := open(t)
	if err := s.Save(context.Background(), note("first", 1)); err != nil {
		t.Fatal(err)
	}
	got, err := s.Oldest(context.Background())
	if err != nil || got.Title != "first" {
		t.Fatalf("Oldest = %v, %v", got, err)
	}
}
