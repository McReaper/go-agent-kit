package m11case

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

type Color int

const (
	Red Color = iota
	Green
	Blue
)

type Job struct {
	ctx context.Context
}

type Payload struct {
	Name string
}

type Guarded struct {
	mu sync.Mutex
	n  int
}

func step() error { return nil }

func Cleanup() {
	os.Remove("a")
	_ = os.Remove("b")
}

func As(v any) int {
	return v.(int)
}

func Show(n int) string {
	return fmt.Sprintf("%d", "n")
}

func Copy(g Guarded) int {
	return g.n
}

func (g *Guarded) Inc() {
	g.mu.Lock()
	g.n++
	g.mu.Unlock()
}

func Swallow() error {
	if err := step(); err != nil {
		return nil
	}
	return nil
}

func Find() (*Payload, error) {
	return nil, nil
}

func Both() error {
	err := step()
	if err != nil {
		return err
	}
	if err2 := step(); err2 != nil {
		return err
	}
	return nil
}

func (j Job) Done() bool {
	return j.ctx.Err() != nil
}

func Tag(ctx context.Context, keys []string) context.Context {
	for _, k := range keys {
		ctx = context.WithValue(ctx, Color(0), k)
	}
	return ctx
}

func Status(ctx context.Context, c *http.Client, url string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return 0, err
	}
	return resp.StatusCode, nil
}

func Request(url string) (*http.Request, error) {
	return http.NewRequest(http.MethodGet, url, nil)
}

func Fetch(url string) (int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func Names(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, "SELECT name FROM notes")
	if err != nil {
		return nil, err
	}
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func Titles(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, "SELECT title FROM notes")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

func Name(c Color) string {
	switch c {
	case Red:
		return "red"
	case Green:
		return "green"
	}
	return ""
}

func Grow(n int) []int {
	xs := make([]int, n)
	xs = append(xs, 1)
	return xs
}

func Wait(d time.Duration) time.Duration {
	return d * time.Second
}

func Hijack() {
	io.EOF = errors.New("hijacked")
}

func Encode(p Payload) ([]byte, error) {
	return json.Marshal(p)
}

func Background(f func()) {
	go f()
}

func Fresh() context.Context {
	return context.Background()
}
