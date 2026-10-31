package list_test

import (
	"context"
	"errors"
	"testing"

	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/domain"
	"github.com/McReaper/go-agent-kit/template/internal/notes/internal/list"
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
		limit *int
		asked int
	}{
		{"20 when none is given", nil, 20},
		{"the one given", new(3), 3},
		{"the smallest", new(1), 1},
		{"100 at most", new(100), 100},
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
	for _, n := range []int{0, -1, 101} {
		if _, err := list.New(&memStore{}).Run(context.Background(), list.Args{Limit: new(n)}); err == nil {
			t.Errorf("limit %d accepted", n)
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
