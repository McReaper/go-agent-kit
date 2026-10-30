// Copyright 2026 Notes Inc. All rights reserved.

package list_test

import (
	"context"
	"errors"
	"testing"

	"example.com/notes/internal/notes/internal/domain"
	"example.com/notes/internal/notes/internal/list"
	"github.com/google/go-cmp/cmp"
)

type memStore struct {
	notes []domain.Note
	asked int
	err   error
}

func (s *memStore) Newest(_ context.Context, limit int) ([]domain.Note, error) {
	s.asked = limit
	return s.notes[:min(limit, len(s.notes))], s.err
}

func TestListAsksTheStoreForTheLimit(t *testing.T) {
	cases := []struct {
		name  string
		limit int
		asked int
	}{
		{"20 when none is given", 0, 20},
		{"the one given", 3, 3},
		{"the smallest", 1, 1},
		{"100 at most", 100, 100},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			store := &memStore{}
			if _, err := list.New(store).Run(context.Background(), list.Args{Limit: c.limit}); err != nil {
				t.Fatal(err)
			}
			if store.asked != c.asked {
				t.Errorf("asked the store for %d, want %d", store.asked, c.asked)
			}
		})
	}
}

func TestListRefusesALimitOutOfRange(t *testing.T) {
	for _, limit := range []int{-1, 101} {
		if _, err := list.New(&memStore{}).Run(context.Background(), list.Args{Limit: limit}); err == nil {
			t.Errorf("limit %d accepted", limit)
		}
	}
}

func TestListReturnsIDsAndTitles(t *testing.T) {
	store := &memStore{notes: []domain.Note{{ID: "a-1", Title: "A", Body: "not listed"}, {ID: "b-2", Title: "B"}}}
	out, err := list.New(store).Run(context.Background(), list.Args{})
	if err != nil {
		t.Fatal(err)
	}
	want := list.Out{Notes: []list.Item{{ID: "a-1", Title: "A"}, {ID: "b-2", Title: "B"}}}
	if diff := cmp.Diff(want, out); diff != "" {
		t.Errorf("Run (-want +got):\n%s", diff)
	}
}

func TestListReportsTheStoreFailure(t *testing.T) {
	boom := errors.New("disk gone")
	_, err := list.New(&memStore{err: boom}).Run(context.Background(), list.Args{})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want it to wrap %v", err, boom)
	}
}
