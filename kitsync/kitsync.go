// Package kitsync brings an app's kit-owned files to a version of the kit:
// the files the template lists in .kit/sync, the section of AGENTS.md
// between the kit's markers, and the kit version kit.yml pins.
package kitsync

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
)

// The template's module, and the paths and markers kit sync knows.
const (
	Template = "github.com/McReaper/go-agent-kit/template"
	Manifest = ".kit/sync"
	Agents   = "AGENTS.md"
	Workflow = ".github/workflows/kit.yml"
	Begin    = "<!-- kit:begin -->"
	End      = "<!-- kit:end -->"
)

// ErrVersion is returned for a version that isn't a full commit id.
var ErrVersion = errors.New("the version is the kit's full commit id: 40 hexadecimal characters")

// Change is one file kit sync wrote or removed: added, updated or removed.
type Change struct {
	Path   string
	Action string
}

var (
	commitID = regexp.MustCompile(`^[0-9a-f]{40}$`)
	pins     = regexp.MustCompile(`(kit-check\.yml@|kit-version: )[0-9a-f]{40}`)
)

// CheckVersion returns ErrVersion unless version is a full commit id.
func CheckVersion(version string) error {
	if !commitID.MatchString(version) {
		return ErrVersion
	}
	return nil
}

// Sync makes the kit-owned files of the app at app what the template at
// template holds, kit.yml pinned to version, and returns what it changed.
func Sync(app, template, version string) (changes []Change, err error) {
	if err := CheckVersion(version); err != nil {
		return nil, err
	}
	src, err := os.OpenRoot(template)
	if err != nil {
		return nil, fmt.Errorf("the template: %w", err)
	}
	defer func() { err = errors.Join(err, src.Close()) }()
	dst, err := os.OpenRoot(app)
	if err != nil {
		return nil, fmt.Errorf("the app: %w", err)
	}
	defer func() { err = errors.Join(err, dst.Close()) }()
	s := &syncer{src: src, dst: dst, version: version}
	if err := s.all(); err != nil {
		return nil, err
	}
	slices.SortFunc(s.changes, func(a, b Change) int { return cmp.Compare(a.Path, b.Path) })
	return s.changes, nil
}

type syncer struct {
	src, dst *os.Root
	version  string
	changes  []Change
}

func (s *syncer) all() error {
	paths, err := s.owned()
	if err != nil {
		return err
	}
	for _, p := range paths {
		if d, ok := strings.CutSuffix(p, "/"); ok {
			err = s.dir(d)
		} else {
			err = s.file(p)
		}
		if err != nil {
			return err
		}
	}
	return s.agents()
}

// WHY: the manifest is the template's, so a kit version that adds or drops
// an owned file says so itself; it is synced like the files it lists.
func (s *syncer) owned() ([]string, error) {
	data, err := s.src.ReadFile(Manifest)
	if err != nil {
		return nil, fmt.Errorf("the template's %s: %w", Manifest, err)
	}
	paths := []string{Manifest}
	for line := range strings.Lines(string(data)) {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			paths = append(paths, line)
		}
	}
	return paths, nil
}

// WHY: a directory the kit owns holds the template's files and only them,
// so a file the kit dropped leaves the app too.
func (s *syncer) dir(d string) error {
	want := map[string]bool{}
	err := fs.WalkDir(s.src.FS(), d, func(p string, e fs.DirEntry, err error) error {
		if err != nil || e.IsDir() {
			return err
		}
		want[p] = true
		return s.file(p)
	})
	if err != nil {
		return fmt.Errorf("the template's %s: %w", d, err)
	}
	return s.prune(d, want)
}

func (s *syncer) prune(d string, want map[string]bool) error {
	err := fs.WalkDir(s.dst.FS(), d, func(p string, e fs.DirEntry, err error) error {
		if errors.Is(err, fs.ErrNotExist) && p == d {
			return nil
		}
		if err != nil || e.IsDir() || want[p] {
			return err
		}
		s.changes = append(s.changes, Change{p, "removed"})
		return s.dst.Remove(p)
	})
	if err != nil {
		return fmt.Errorf("the app's %s: %w", d, err)
	}
	return nil
}

func (s *syncer) file(p string) error {
	data, err := s.src.ReadFile(p)
	if err != nil {
		return fmt.Errorf("the template's %s: %w", p, err)
	}
	if p == Workflow {
		data = pins.ReplaceAll(data, []byte("${1}"+s.version))
	}
	return s.write(p, data)
}

func (s *syncer) write(p string, data []byte) error {
	info, err := s.src.Stat(p)
	if err != nil {
		return fmt.Errorf("the template's %s: %w", p, err)
	}
	old, err := s.dst.ReadFile(p)
	switch {
	case err == nil && bytes.Equal(old, data):
		return nil
	case err == nil:
		s.changes = append(s.changes, Change{p, "updated"})
	case errors.Is(err, fs.ErrNotExist):
		s.changes = append(s.changes, Change{p, "added"})
	default:
		return fmt.Errorf("the app's %s: %w", p, err)
	}
	if err := s.dst.MkdirAll(path.Dir(p), 0o750); err != nil {
		return fmt.Errorf("the app's %s: %w", path.Dir(p), err)
	}
	if err := s.dst.WriteFile(p, data, info.Mode().Perm()); err != nil {
		return fmt.Errorf("the app's %s: %w", p, err)
	}
	return s.dst.Chmod(p, info.Mode().Perm())
}

// WHY: AGENTS.md is the app's, but for its section between the kit's
// markers: the commands and the rules every app shares. An app without the
// markers gets the section at its end, and syncs in place from then on.
func (s *syncer) agents() error {
	tmpl, err := s.src.ReadFile(Agents)
	if err != nil {
		return fmt.Errorf("the template's %s: %w", Agents, err)
	}
	section, ok := between(string(tmpl))
	if !ok {
		return fmt.Errorf("the template's %s has no section between %s and %s", Agents, Begin, End)
	}
	app, err := s.dst.ReadFile(Agents)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("the app's %s: %w", Agents, err)
	}
	return s.write(Agents, []byte(withSection(string(app), section)))
}

func between(text string) (string, bool) {
	i, j := strings.Index(text, Begin), strings.Index(text, End)
	if i < 0 || j < i {
		return "", false
	}
	return text[i : j+len(End)], true
}

func withSection(app, section string) string {
	if old, ok := between(app); ok {
		return strings.Replace(app, old, section, 1)
	}
	if app == "" {
		return section + "\n"
	}
	return strings.TrimRight(app, "\n") + "\n\n" + section + "\n"
}
