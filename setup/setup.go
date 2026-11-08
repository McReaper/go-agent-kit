// Package setup sets a repository up for Rumblestrip. It writes the few files
// the repository keeps: the CI that calls the merge check, the code owners
// of the harness, the bot that moves Rumblestrip's pin, and the Claude Code
// settings that turn Rumblestrip's plugin on. Everything else (the checks, the
// rules, the skills, the hooks) stays in Rumblestrip, at the version the CI
// pins. As the owner, it also applies the GitHub settings that make the
// merge check the arbiter, and creates the agent's GitHub App.
package setup

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"regexp"
	"text/template"

	"github.com/McReaper/rumblestrip-go/guard"
)

//go:embed files
var files embed.FS

// Version is the Rumblestrip release a repository pins: its tag, and the commit
// the tag names, which the CI pins by.
type Version struct {
	Tag    string
	Commit string
}

// Options say what the files name: the Rumblestrip version, the owner's login, and
// the owner's address the agent's commits name them co-author by.
type Options struct {
	Version    Version
	Owner      string
	OwnerEmail string
}

// File is one file of the repository: its path from the root, its content.
type File struct {
	Path    string
	Content []byte
}

// Files returns the files a repository keeps for Rumblestrip, for o.
func Files(o Options) ([]File, error) {
	if o.Owner == "" || o.OwnerEmail == "" || len(o.Version.Commit) != 40 {
		return nil, errors.New("rumblestrip init needs the owner's login and address and Rumblestrip's full commit id")
	}
	return renderAll(o, []struct{ path, from string }{
		{".github/workflows/rumblestrip.yml", "rumblestrip.yml"},
		{".github/workflows/rumblestrip-approval.yml", "rumblestrip-approval.yml"},
		{".github/CODEOWNERS", "CODEOWNERS"},
		{".github/dependabot.yml", "dependabot.yml"},
		{".claude/settings.json", "settings.json"},
	})
}

type templateData struct {
	Version
	Owner      string
	OwnerEmail string
	Harness    []string
}

func renderAll(o Options, fs []struct{ path, from string }) ([]File, error) {
	data := templateData{o.Version, o.Owner, o.OwnerEmail, codeOwned()}
	var out []File
	for _, f := range fs {
		content, err := render(f.from, data)
		if err != nil {
			return nil, err
		}
		out = append(out, File{f.path, content})
	}
	return out, nil
}

// WHY: CODEOWNERS names the guard's harness, so GitHub and Rumblestrip's checks
// hold the same paths.
func codeOwned() []string {
	var out []string
	for _, h := range guard.Harness() {
		if h == "GLOSSARY.md" {
			out = append(out, "**/"+h)
		} else {
			out = append(out, "/"+h)
		}
	}
	return out
}

func render(name string, data any) ([]byte, error) {
	text, err := files.ReadFile("files/" + name)
	if err != nil {
		return nil, fmt.Errorf("rumblestrip init: %w", err)
	}
	t, err := template.New(name).Option("missingkey=error").Parse(string(text))
	if err != nil {
		return nil, fmt.Errorf("rumblestrip init: %s: %w", name, err)
	}
	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		return nil, fmt.Errorf("rumblestrip init: %s: %w", name, err)
	}
	return b.Bytes(), nil
}

// Report says what Write did with each file: written, already there as the
// rumblestrip writes it, or there with other content, which it keeps.
type Report struct {
	Written, Same, Kept []string
}

// Write writes each file into the repository at dir, unless it is there.
//
// WHY: a file the repository already has may hold the owner's own changes
// (a second code owner): rumblestrip init lists it rather than overwriting it, and
// the owner decides.
func Write(dir string, fs []File) (r Report, err error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return Report{}, fmt.Errorf("rumblestrip init: %w", err)
	}
	defer func() { err = errors.Join(err, root.Close()) }()
	for _, f := range fs {
		if err := r.write(root, f); err != nil {
			return r, err
		}
	}
	return r, nil
}

func (r *Report) write(root *os.Root, f File) error {
	old, err := root.ReadFile(f.Path)
	switch {
	case err == nil && bytes.Equal(old, f.Content):
		r.Same = append(r.Same, f.Path)
		return nil
	case err == nil:
		r.Kept = append(r.Kept, f.Path)
		return nil
	case !errors.Is(err, fs.ErrNotExist):
		return fmt.Errorf("rumblestrip init: %w", err)
	}
	if err := root.MkdirAll(path.Dir(f.Path), 0o750); err != nil {
		return fmt.Errorf("rumblestrip init: %w", err)
	}
	if err := root.WriteFile(f.Path, f.Content, 0o600); err != nil {
		return fmt.Errorf("rumblestrip init: %w", err)
	}
	r.Written = append(r.Written, f.Path)
	return nil
}

var (
	pin       = regexp.MustCompile(`rumblestrip-check\.yml@([0-9a-f]{40}) # (\S+)`)
	ownerLine = regexp.MustCompile(`(?m)^      owner: (\S+)$`)
)

// WorkflowsFor returns the workflows rumblestrip init writes for the pin and the
// owner kitYML, a repository's rumblestrip.yml, holds: what its rumblestrip.yml and
// rumblestrip-approval.yml must be.
//
// WHY: a Rumblestrip version may need other workflows, a trigger or a permission
// more; the repository's, pinned to that version, can't say so itself.
func WorkflowsFor(kitYML []byte) ([]File, error) {
	m := pin.FindSubmatch(kitYML)
	if m == nil {
		return nil, errors.New("rumblestrip.yml pins no rumblestrip commit and release (rumblestrip-check.yml@<commit> # <tag>)")
	}
	owner := ownerLine.FindSubmatch(kitYML)
	if owner == nil {
		return nil, errors.New("rumblestrip.yml names no owner (owner: <login>, under with:)")
	}
	return renderAll(Options{Version: Version{Tag: string(m[2]), Commit: string(m[1])}, Owner: string(owner[1])}, []struct{ path, from string }{
		{".github/workflows/rumblestrip.yml", "rumblestrip.yml"},
		{".github/workflows/rumblestrip-approval.yml", "rumblestrip-approval.yml"},
	})
}
