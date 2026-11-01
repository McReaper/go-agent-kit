package mutate

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/McReaper/go-agent-kit/internal/platform/git"
)

// WHY: the copy holds the files git sees in dir, tracked and new; git's
// list leaves out what the repository ignores: build outputs,
// databases and binaries that can weigh gigabytes and that tests don't read.
func copyModule(ctx context.Context, dir string) (work string, err error) {
	out, err := git.Run(ctx, dir, "ls-files", "-z", "--cached", "--others", "--exclude-standard")
	if err != nil {
		return "", err
	}
	work, err = os.MkdirTemp("", "kit-mutate-*")
	if err != nil {
		return "", fmt.Errorf("make the copy's directory: %w", err)
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, os.RemoveAll(work))
		}
	}()
	return work, copyFiles(dir, work, strings.Split(strings.TrimSuffix(out, "\x00"), "\x00"))
}

func copyFiles(from, to string, files []string) (err error) {
	src, err := os.OpenRoot(from)
	if err != nil {
		return fmt.Errorf("open %s: %w", from, err)
	}
	defer func() { err = errors.Join(err, src.Close()) }()
	dst, err := os.OpenRoot(to)
	if err != nil {
		return fmt.Errorf("open %s: %w", to, err)
	}
	defer func() { err = errors.Join(err, dst.Close()) }()
	for _, p := range files {
		if err := copyFile(src, dst, p); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst *os.Root, p string) error {
	if p == "" {
		return nil
	}
	data, err := src.ReadFile(p)
	// WHY: a file deleted in the working tree is still in git's index.
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read %s: %w", p, err)
	}
	if err := dst.MkdirAll(path.Dir(p), 0o750); err != nil {
		return fmt.Errorf("copy %s: %w", p, err)
	}
	if err := dst.WriteFile(p, data, 0o600); err != nil {
		return fmt.Errorf("copy %s: %w", p, err)
	}
	return nil
}
