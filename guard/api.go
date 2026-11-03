package guard

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

	"github.com/McReaper/go-agent-kit/internal/platform/run"
)

// WHY: only pkg/ holds libraries other programs import, and only an
// incompatible change breaks them: that one is the owner's call (E5). An
// app's contract with its users is its spec.
func apiChanges(ctx context.Context, dir string, d Diff) (items []Finding, err error) {
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
	t := trees{head: dir, base: old, module: module}
	if err := t.extractBase(ctx, d.Base); err != nil {
		return nil, err
	}
	for _, p := range pkgs {
		more, err := t.packageAPI(ctx, p)
		if err != nil {
			return nil, err
		}
		items = append(items, more...)
	}
	return items, nil
}

func libraryPackages(d Diff) []string {
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

// WHY: the change's module at head, its base's tree extracted beside it,
// and the module's path, which apidiff names packages by.
type trees struct {
	head, base, module string
}

func (t trees) packageAPI(ctx context.Context, p string) ([]Finding, error) {
	before, after := hasGo(filepath.Join(t.base, p)), hasGo(filepath.Join(t.head, p))
	switch {
	case !before:
		return nil, nil
	case !after:
		return []Finding{apiBreak(p, "the package is removed")}, nil
	}
	export := filepath.Join(t.base, ".kit-api-export")
	if _, err := output(ctx, t.base, "apidiff", "-w", export, t.module+"/"+p); err != nil {
		return nil, err
	}
	out, err := output(ctx, t.head, "apidiff", export, t.module+"/"+p)
	if err != nil {
		return nil, err
	}
	return apidiffItems(p, out), nil
}

func apidiffItems(p, out string) []Finding {
	var items []Finding
	incompatible := false
	for line := range strings.Lines(out) {
		line = strings.TrimSpace(line)
		switch {
		case line == "Incompatible changes:" || line == "Compatible changes:":
			incompatible = line == "Incompatible changes:"
		case strings.HasPrefix(line, "- ") && incompatible:
			items = append(items, apiBreak(p, strings.TrimPrefix(line, "- ")))
		}
	}
	return items
}

func apiBreak(p, what string) Finding {
	return Finding{
		"protected-api", p, 0,
		"E5: an incompatible change to the library's API, " + what + ": every program that imports it breaks; the owner approves it",
	}
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
	out, err := run.Command(ctx, dir, argv...)
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
func (t trees) extractBase(ctx context.Context, base string) (err error) {
	out, err := run.Command(ctx, t.head, "git", "archive", "--format=tar", base)
	if err != nil {
		return err
	}
	if out.Code != 0 {
		return fmt.Errorf("git archive %s: exit %d: %s", base, out.Code, strings.TrimSpace(string(out.Stderr)))
	}
	root, err := os.OpenRoot(t.base)
	if err != nil {
		return fmt.Errorf("open %s: %w", t.base, err)
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
