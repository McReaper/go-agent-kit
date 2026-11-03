// Package gesture refuses the agent's gestures that reach past the kit's
// judgement: merging, pushing to the default branch, a force push that
// could rewrite it, a commit that skips its hooks, a change to the rules,
// and an edit of what runs or configures the checks. Claude Code asks it
// before each command and edit (kit hook pre-tool); the platform still
// refuses what slips past, this only says so before the agent tries.
package gesture

import (
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// Tool is what the agent is about to do, as Claude Code's PreToolUse hook
// gives it: the tool's name and its input.
type Tool struct {
	Name     string
	Command  string // Bash's command line
	FilePath string // the file an edit writes
	Dir      string // where the agent works, the repository's root or inside it
}

// Refusal says why a gesture is refused and what to do instead.
type Refusal string

// Check returns why t is refused, or "" when it is allowed.
func Check(t Tool) Refusal {
	switch t.Name {
	case "Bash":
		return bash(t.Command)
	case "Edit", "Write", "MultiEdit", "NotebookEdit":
		return edit(t.Dir, t.FilePath)
	}
	return ""
}

func bash(line string) Refusal {
	for _, c := range commands(line) {
		if r := command(c); r != "" {
			return r
		}
	}
	return ""
}

var separator = regexp.MustCompile(`&&|\|\||[;|\n]`)

// WHY: a deny rule matches the command's text, so `git -C . push` slipped
// past `git push`; the words are read after git's own options.
func commands(line string) [][]string {
	var out [][]string
	for part := range strings.SplitSeq(separator.ReplaceAllString(line, "\x00"), "\x00") {
		if w := strings.Fields(part); len(w) > 0 {
			out = append(out, withoutGitOptions(w))
		}
	}
	return out
}

func withoutGitOptions(w []string) []string {
	if path.Base(w[0]) != "git" && path.Base(w[0]) != "git.exe" {
		return w
	}
	out := []string{"git"}
	i := 1
	for i < len(w) && strings.HasPrefix(w[i], "-") {
		if w[i] == "-C" || w[i] == "-c" {
			i++
		}
		i++
	}
	return append(out, w[min(i, len(w)):]...)
}

func command(w []string) Refusal {
	switch {
	case len(w) >= 3 && w[0] == "gh" && w[1] == "pr" && w[2] == "merge":
		return "T2: a pull request merges by itself once its checks pass and the owner approved what is theirs: never merge it"
	case len(w) >= 2 && w[0] == "gh" && w[1] == "api" && slices.ContainsFunc(w, func(a string) bool { return strings.Contains(a, "rulesets") }):
		return "T2: the rulesets are the owner's: never change them"
	case len(w) >= 2 && w[0] == "git" && w[1] == "commit" && (slices.Contains(w, "--no-verify") || slices.Contains(w, "-n")):
		return "T2: a commit runs its hooks: fix what they report instead of skipping them"
	case len(w) >= 2 && w[0] == "git" && w[1] == "push":
		return push(w[2:])
	}
	return ""
}

func push(args []string) Refusal {
	for _, a := range args {
		switch {
		case a == "--force" || a == "-f" || strings.HasPrefix(a, "+"):
			return "T2: a force push rewrites what others built on: push with --force-with-lease, to your own branch"
		case a == "main" || strings.HasSuffix(a, ":main") || a == "master" || strings.HasSuffix(a, ":master"):
			return "T2: the default branch changes through a pull request only: push your branch and open one"
		}
	}
	return ""
}

// WHY: what runs the checks and configures the agent is the owner's: CI,
// the kit's settings, the agent's settings. A change to AGENTS.md or an ADR
// goes through a pull request the owner approves, so those stay editable.
func edit(dir, file string) Refusal {
	rel := file
	if filepath.IsAbs(file) && dir != "" {
		r, err := filepath.Rel(dir, file)
		if err != nil || strings.HasPrefix(r, "..") {
			return ""
		}
		rel = r
	}
	rel = filepath.ToSlash(rel)
	for _, p := range []string{".github/", ".kit/", ".claude/"} {
		if strings.HasPrefix(rel, p) {
			return Refusal("T2: " + p + " is what runs the checks or configures the agent, the owner's to change: say in your report what should change there and why")
		}
	}
	return ""
}

// Commits reports whether line commits: the moment to stop a bypass or a
// secret before it enters the history.
func Commits(line string) bool {
	for _, w := range commands(line) {
		if len(w) >= 2 && w[0] == "git" && w[1] == "commit" {
			return true
		}
	}
	return false
}
