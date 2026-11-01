package guard

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/McReaper/go-agent-kit/internal/platform/git"
)

// Tree is an approved commit's files: each path's git blob id.
type Tree map[string]string

// Approval is what a code owner approved on a change: its head, or earlier
// commits whose trees hold the protected files they saw.
type Approval struct {
	Head  bool
	Trees []Tree
}

// Covered returns the findings an approval is enough for. A protected file
// is covered while its content at HEAD is its content in an approved tree:
// the agent goes on without asking again, and touching the file reopens the
// approval. A change over the size needs the head approved: more code is
// more to read.
func Covered(ctx context.Context, dir string, fs []Finding, a Approval) ([]Finding, error) {
	var out []Finding
	for _, f := range fs {
		ok, err := covers(ctx, dir, f, a)
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, f)
		}
	}
	return out, nil
}

func covers(ctx context.Context, dir string, f Finding, a Approval) (bool, error) {
	switch {
	case !f.NeedsApproval():
		return false, nil
	case a.Head:
		return true, nil
	case f.File == "" || strings.HasPrefix(f.Rule, "size-") || len(a.Trees) == 0:
		return false, nil
	}
	head, err := blob(ctx, dir, f.File)
	if err != nil {
		return false, err
	}
	for _, t := range a.Trees {
		if t[f.File] == head {
			return true, nil
		}
	}
	return false, nil
}

// WHY: an empty id stands for a file HEAD doesn't have, so an approved
// deletion stays approved while the file stays deleted.
func blob(ctx context.Context, dir, path string) (string, error) {
	out, err := git.Run(ctx, dir, "ls-tree", "HEAD", "--", path)
	if err != nil {
		return "", fmt.Errorf("read %s at HEAD: %w", path, err)
	}
	fields := strings.Fields(out)
	if len(fields) < 3 {
		return "", nil
	}
	return fields[2], nil
}

// ReadTrees reads the approved trees in dir, one file per approved commit,
// each line a blob id, a space and a path. No dir, or one that doesn't
// exist, holds none.
func ReadTrees(dir string) (trees []Tree, err error) {
	if dir == "" {
		return nil, nil
	}
	root, err := os.OpenRoot(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("approved trees: %w", err)
	}
	defer func() { err = errors.Join(err, root.Close()) }()
	entries, err := fs.ReadDir(root.FS(), ".")
	if err != nil {
		return nil, fmt.Errorf("approved trees: %w", err)
	}
	for _, e := range entries {
		data, err := root.ReadFile(e.Name())
		if err != nil {
			return nil, fmt.Errorf("approved tree %s: %w", e.Name(), err)
		}
		t, err := ParseTree(string(data))
		if err != nil {
			return nil, fmt.Errorf("approved tree %s: %w", e.Name(), err)
		}
		trees = append(trees, t)
	}
	return trees, nil
}

// ParseTree reads one approved tree: each line a blob id, a space and a path.
func ParseTree(text string) (Tree, error) {
	t := Tree{}
	sc := bufio.NewScanner(strings.NewReader(text))
	for sc.Scan() {
		id, path, ok := strings.Cut(sc.Text(), " ")
		if !ok || id == "" || path == "" {
			return nil, fmt.Errorf("%q is not a blob id and a path", sc.Text())
		}
		t[path] = id
	}
	return t, sc.Err()
}
