package review

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/McReaper/go-agent-kit/guard"
	"github.com/McReaper/go-agent-kit/internal/platform/run"
)

// WHY: only pkg/ holds libraries other programs import; an app's contract
// with its users is its golden snapshot, listed apart.
func apiChanges(ctx context.Context, dir string, d guard.Diff) (items []Item, err error) {
	pkgs := libraryPackages(d)
	if len(pkgs) == 0 {
		return nil, nil
	}
	module, err := output(ctx, dir, "go", "list", "-m")
	if err != nil {
		return nil, err
	}
	old, err := os.MkdirTemp("", "kit-review-*")
	if err != nil {
		return nil, fmt.Errorf("make the base's directory: %w", err)
	}
	defer func() { err = errors.Join(err, os.RemoveAll(old)) }()
	if err := extractBase(ctx, dir, d.Base, old); err != nil {
		return nil, err
	}
	for _, p := range pkgs {
		more, err := packageAPI(ctx, trees{head: dir, base: old}, module, p)
		if err != nil {
			return nil, err
		}
		items = append(items, more...)
	}
	return items, nil
}

func libraryPackages(d guard.Diff) []string {
	var out []string
	for _, f := range d.Files {
		p := path.Dir(f.Path)
		if strings.HasPrefix(f.Path, "pkg/") && path.Ext(f.Path) == ".go" && !strings.HasSuffix(f.Path, "_test.go") && !slices.Contains(out, p) {
			out = append(out, p)
		}
	}
	slices.Sort(out)
	return out
}

type trees struct {
	head, base string
}

func packageAPI(ctx context.Context, t trees, module, p string) ([]Item, error) {
	before, after := hasGo(filepath.Join(t.base, p)), hasGo(filepath.Join(t.head, p))
	switch {
	case !before && !after:
		return nil, nil
	case !before:
		return []Item{{File: p, Text: "new package"}}, nil
	case !after:
		return []Item{{File: p, Text: "package removed"}}, nil
	}
	export := filepath.Join(t.base, ".kit-api-export")
	if _, err := output(ctx, t.base, "apidiff", "-w", export, module+"/"+p); err != nil {
		return nil, err
	}
	out, err := output(ctx, t.head, "apidiff", export, module+"/"+p)
	if err != nil {
		return nil, err
	}
	return apidiffItems(p, out), nil
}

func apidiffItems(p, out string) []Item {
	var items []Item
	kind := ""
	for line := range strings.Lines(out) {
		line = strings.TrimSpace(line)
		switch {
		case line == "Incompatible changes:":
			kind = "incompatible"
		case line == "Compatible changes:":
			kind = "compatible"
		case strings.HasPrefix(line, "- ") && kind != "":
			items = append(items, Item{File: p, Text: kind + ": " + strings.TrimPrefix(line, "- ")})
		}
	}
	return items
}

func hasGo(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	return slices.ContainsFunc(entries, func(e os.DirEntry) bool {
		return !e.IsDir() && path.Ext(e.Name()) == ".go" && !strings.HasSuffix(e.Name(), "_test.go")
	})
}

func output(ctx context.Context, dir string, argv ...string) (string, error) {
	out, err := run.Command(ctx, dir, nil, argv...)
	if err != nil {
		return "", err
	}
	if out.Code != 0 {
		return "", fmt.Errorf("%s: exit %d: %s", strings.Join(argv, " "), out.Code, strings.TrimSpace(string(out.Stderr)))
	}
	return strings.TrimSpace(string(out.Stdout)), nil
}

// WHY: git archive writes the base's tree without touching the repository:
// a worktree would add to .git while the agent works in it.
func extractBase(ctx context.Context, dir, base, to string) (err error) {
	out, err := run.Command(ctx, dir, nil, "git", "archive", "--format=tar", base)
	if err != nil {
		return err
	}
	if out.Code != 0 {
		return fmt.Errorf("git archive %s: exit %d: %s", base, out.Code, strings.TrimSpace(string(out.Stderr)))
	}
	root, err := os.OpenRoot(to)
	if err != nil {
		return fmt.Errorf("open %s: %w", to, err)
	}
	defer func() { err = errors.Join(err, root.Close()) }()
	tr := tar.NewReader(bytes.NewReader(out.Stdout))
	for {
		h, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read the base's archive: %w", err)
		}
		if err := extract(root, h, tr); err != nil {
			return err
		}
	}
}

func extract(root *os.Root, h *tar.Header, r io.Reader) error {
	switch h.Typeflag {
	case tar.TypeDir:
		if err := root.MkdirAll(h.Name, 0o750); err != nil {
			return fmt.Errorf("extract %s: %w", h.Name, err)
		}
	case tar.TypeReg:
		data, err := io.ReadAll(r)
		if err != nil {
			return fmt.Errorf("extract %s: %w", h.Name, err)
		}
		if err := root.MkdirAll(path.Dir(h.Name), 0o750); err != nil {
			return fmt.Errorf("extract %s: %w", h.Name, err)
		}
		if err := root.WriteFile(h.Name, data, 0o600); err != nil {
			return fmt.Errorf("extract %s: %w", h.Name, err)
		}
	}
	return nil
}
