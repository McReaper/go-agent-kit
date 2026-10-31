// Package guard checks a change against the rules no linter can see: what
// it adds that hides code from the checks, which protected files it
// touches, and how big it is (criteria T2 and T3).
package guard

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/McReaper/go-agent-kit/internal/platform/git"
)

// Line is one added line: its number in the new file and its text.
type Line struct {
	N    int
	Text string
}

// File is one file the change touches.
type File struct {
	Path    string
	OldPath string // the path before a rename, else empty
	Status  byte   // 'A' added, 'M' modified, 'D' deleted, 'R' renamed
	Added   []Line
	Removed int
}

// Diff is a change: every file it touches, and the commit it starts from.
type Diff struct {
	Files []File
	Base  string // base's merge base with HEAD, set by Read
}

// Read returns the change from base's merge base to the working tree of
// the git repository at dir: commits, staged and unstaged edits, and new
// files git doesn't track yet.
//
// WHY: an agent runs the checks before it commits; a guard that read
// commits only passed work it never saw.
func Read(ctx context.Context, dir, base string) (Diff, error) {
	if err := git.CheckRef(base); err != nil {
		return Diff{}, fmt.Errorf("base: %w", err)
	}
	mb, err := git.Run(ctx, dir, "merge-base", base, "HEAD")
	if err != nil {
		return Diff{}, err
	}
	d, err := readTracked(ctx, dir, strings.TrimSpace(mb))
	if err != nil {
		return Diff{}, err
	}
	untracked, err := readUntracked(ctx, dir)
	if err != nil {
		return Diff{}, err
	}
	d.Files = append(d.Files, untracked...)
	d.Base = strings.TrimSpace(mb)
	return d, nil
}

func readTracked(ctx context.Context, dir, from string) (Diff, error) {
	status, err := git.Run(ctx, dir, "diff", "--name-status", "-M", from)
	if err != nil {
		return Diff{}, err
	}
	numstat, err := git.Run(ctx, dir, "diff", "--numstat", "-M", from)
	if err != nil {
		return Diff{}, err
	}
	patch, err := git.Run(ctx, dir, "diff", "-U0", "--no-color", "--no-prefix", "-M", from)
	if err != nil {
		return Diff{}, err
	}
	return Parse(status, numstat, patch)
}

func readUntracked(ctx context.Context, dir string) (files []File, err error) {
	out, err := git.Run(ctx, dir, "ls-files", "-z", "--others", "--exclude-standard")
	if err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", dir, err)
	}
	defer func() { err = errors.Join(err, root.Close()) }()
	for p := range strings.SplitSeq(strings.TrimSuffix(out, "\x00"), "\x00") {
		if p == "" {
			continue
		}
		data, err := root.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", p, err)
		}
		files = append(files, File{Path: p, Status: 'A', Added: linesOf(string(data))})
	}
	return files, nil
}

func linesOf(text string) []Line {
	var out []Line
	for l := range strings.Lines(text) {
		out = append(out, Line{N: len(out) + 1, Text: strings.TrimRight(l, "\r\n")})
	}
	return out
}

// Parse builds a Diff from the output of git diff --name-status, --numstat
// and -U0 --no-prefix, all with -M.
func Parse(status, numstat, patch string) (Diff, error) {
	files, err := parseStatus(status)
	if err != nil {
		return Diff{}, err
	}
	removed, err := parseNumstat(numstat)
	if err != nil {
		return Diff{}, err
	}
	added, err := parsePatch(patch)
	if err != nil {
		return Diff{}, err
	}
	for i := range files {
		files[i].Added = added[files[i].Path]
		files[i].Removed = removed[files[i].Path]
	}
	return Diff{Files: files}, nil
}

func parseStatus(out string) ([]File, error) {
	var files []File
	for line := range strings.Lines(out) {
		fields := strings.Split(strings.TrimRight(line, "\r\n"), "\t")
		if len(fields) < 2 || fields[0] == "" {
			return nil, fmt.Errorf("git diff --name-status: unexpected line %q", line)
		}
		f := File{Status: fields[0][0], Path: fields[len(fields)-1]}
		if f.Status == 'R' && len(fields) == 3 {
			f.OldPath = fields[1]
		}
		files = append(files, f)
	}
	return files, nil
}

func parseNumstat(out string) (map[string]int, error) {
	removed := map[string]int{}
	for line := range strings.Lines(out) {
		fields := strings.Split(strings.TrimRight(line, "\r\n"), "\t")
		if len(fields) != 3 {
			return nil, fmt.Errorf("git diff --numstat: unexpected line %q", line)
		}
		n, err := count(fields[1])
		if err != nil {
			return nil, fmt.Errorf("git diff --numstat: %w", err)
		}
		removed[renamedTo(fields[2])] = n
	}
	return removed, nil
}

// WHY: numstat prints "-" for a binary file, whose lines don't count.
func count(field string) (int, error) {
	if field == "-" {
		return 0, nil
	}
	n, err := strconv.Atoi(field)
	if err != nil {
		return 0, fmt.Errorf("line count %q: %w", field, err)
	}
	return n, nil
}

// WHY: numstat names a rename "dir/{old => new}/f" or "old => new"; the
// rules want the new path, as name-status gives it.
func renamedTo(path string) string {
	before, rest, ok := strings.Cut(path, "{")
	if !ok {
		_, after, renamed := strings.Cut(path, " => ")
		if renamed {
			return after
		}
		return path
	}
	inner, after, _ := strings.Cut(rest, "}")
	_, to, _ := strings.Cut(inner, " => ")
	return strings.ReplaceAll(before+to+after, "//", "/")
}

type patchReader struct {
	added map[string][]Line
	file  string
	next  int
}

func parsePatch(out string) (map[string][]Line, error) {
	r := patchReader{added: map[string][]Line{}}
	sc := bufio.NewScanner(strings.NewReader(out))
	sc.Buffer(make([]byte, 1<<20), 1<<24)
	for sc.Scan() {
		if err := r.line(sc.Text()); err != nil {
			return nil, err
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("git diff: %w", err)
	}
	return r.added, nil
}

func (r *patchReader) line(line string) error {
	switch {
	case strings.HasPrefix(line, "+++ "):
		r.file, r.next = strings.TrimPrefix(line, "+++ "), 0
	case strings.HasPrefix(line, "@@ "):
		n, err := hunkStart(line)
		r.next = n
		return err
	case strings.HasPrefix(line, "+") && r.file != "/dev/null":
		r.added[r.file] = append(r.added[r.file], Line{N: r.next, Text: line[1:]})
		r.next++
	}
	return nil
}

func hunkStart(header string) (int, error) {
	_, plus, _ := strings.Cut(header, " +")
	start, _, _ := strings.Cut(plus, " ")
	start, _, _ = strings.Cut(start, ",")
	n, err := strconv.Atoi(start)
	if err != nil {
		return 0, fmt.Errorf("git diff: hunk header %q: %w", header, err)
	}
	return n, nil
}
