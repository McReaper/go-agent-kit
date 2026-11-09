package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/McReaper/rumblestrip-go/check"
)

// WHY: GitHub shows ten error annotations a step and drops the rest, and
// a summary table of hundreds of rows is read by no one: the log keeps
// every finding.
const (
	maxAnnotations = 10
	maxRows        = 50
	maxMessage     = 300
)

var findingAt = regexp.MustCompile(`^([^\s:]+):(\d+):(?:\d+:)? ?(.*)$`)

// WHY: on GitHub a finding is read in two places besides the log: on its
// line, in the pull request's files, and in the run's one summary, where
// the owner reads what each check is for, then what a failing one wants
// and whose the finding is to fix. Check writes that summary from what each
// gate's job leaves in summaryDir: a row, the gate's title and what it
// checks, and its section when it failed.
func reportGitHub(out io.Writer, summaryDir string, results []check.Result) error {
	lines := annotations(results)
	if _, err := io.WriteString(out, strings.Join(lines[:min(len(lines), maxAnnotations)], "")); err != nil {
		return fmt.Errorf("annotate the findings: %w", err)
	}
	if summaryDir == "" {
		return nil
	}
	return writeSummary(summaryDir, results)
}

func writeSummary(dir string, results []check.Result) (err error) {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("write the summary: %w", err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return fmt.Errorf("write the summary: %w", err)
	}
	defer func() { err = errors.Join(err, root.Close()) }()
	for _, r := range results {
		if err := writeGateSummary(root, r); err != nil {
			return err
		}
	}
	return nil
}

func writeGateSummary(root *os.Root, r check.Result) error {
	if err := root.WriteFile(r.ID+".row", []byte(r.Title+"\t"+r.Purpose+"\n"), 0o600); err != nil {
		return fmt.Errorf("write the summary: %w", err)
	}
	s := gateSection(r)
	if s == "" {
		return nil
	}
	if err := root.WriteFile(r.ID+".md", []byte(s), 0o600); err != nil {
		return fmt.Errorf("write the summary: %w", err)
	}
	return nil
}

func annotations(results []check.Result) []string {
	var out []string
	for _, r := range results {
		out = append(out, gateAnnotations(r)...)
	}
	return out
}

func gateAnnotations(r check.Result) []string {
	var out []string
	for _, f := range r.Findings {
		if a, ok := annotation(r.Title, f); ok {
			out = append(out, a)
		}
	}
	return out
}

// WHY: GitHub pins an annotation with no file to the workflow's file, where
// it reads as the workflow's fault; a finding with no line is the
// summary's and the log's.
func annotation(title, finding string) (string, bool) {
	m := findingAt.FindStringSubmatch(finding)
	if m == nil {
		return "", false
	}
	return "::error file=" + escapeProperty(m[1]) + ",line=" + m[2] + ",title=" + escapeProperty(title) + "::" + escapeData(shorten(forPeople(m[3]))) + "\n", true
}

var (
	ruleFirst = regexp.MustCompile(`^[a-z][a-z0-9-]*: (?:[A-Z]\d*: )?`)
	subtest   = regexp.MustCompile(`\bTest\w+(?:/[^\s:]+)+`)
)

// WHY: the log's line leads with the rule and the criterion, for the agent
// to look up; a person reads the sentence. Go writes a subtest's name with
// underscores for spaces: a person reads the words.
func forPeople(s string) string {
	s = ruleFirst.ReplaceAllString(s, "")
	return subtest.ReplaceAllStringFunc(s, func(name string) string {
		parts := strings.Split(name, "/")
		for i := 1; i < len(parts); i++ {
			parts[i] = strings.ReplaceAll(parts[i], "_", " ")
		}
		return strings.Join(parts, " › ")
	})
}

// WHY: a test's message can quote a whole program's output; the summary and
// the annotation say enough to know what broke, and the log keeps it all.
// The cut is in the middle: a message says what happened first and what
// was wanted last.
func shorten(s string) string {
	s = oneLine(s)
	runes := []rune(s)
	if len(runes) <= maxMessage {
		return s
	}
	half := maxMessage / 2
	return string(runes[:half]) + " … " + string(runes[len(runes)-half:])
}

func escapeData(s string) string {
	return strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A").Replace(s)
}

func escapeProperty(s string) string {
	return strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A", ":", "%3A", ",", "%2C").Replace(s)
}

func gateSection(r check.Result) string {
	var b strings.Builder
	switch {
	case r.Err != nil:
		fmt.Fprintf(&b, "### ❌ %s\n\nIt could not run: `%s`. That is Rumblestrip's or the runner's fault, not the change's: re-run the job.\n\n", r.Title, oneLine(r.Err.Error()))
	case len(r.Findings) > 0:
		fmt.Fprintf(&b, "### ❌ %s\n\nIt wants: %s It found %s, for the agent to fix:\n\n", r.Title, lowerFirst(r.Purpose), count(len(r.Findings), "problem"))
		writeRows(&b, r.Findings)
	case len(r.Notes) > 0:
		fmt.Fprintf(&b, "### ✅ %s\n\n%s\n\n", r.Title, r.Purpose)
	}
	if len(r.Notes) > 0 {
		fmt.Fprintf(&b, "%s yours to approve: Rumblestrip's comment on the pull request lists them.\n\n", count(len(r.Notes), "change is", "changes are"))
	}
	return b.String()
}

func writeRows(b *strings.Builder, findings []string) {
	b.WriteString("| Where | What |\n| --- | --- |\n")
	for i, f := range findings {
		if i == maxRows {
			fmt.Fprintf(b, "\nAnd %d more, in the job's log.\n", len(findings)-maxRows)
			break
		}
		where, what := place(f)
		fmt.Fprintf(b, "| %s | %s |\n", where, cell(forPeople(what)))
	}
	b.WriteString("\n")
}

// WHY: a finding names a file and a line, or, when what failed lies outside
// the module (a test library's helper), the package it failed in.
func place(f string) (where, what string) {
	if m := findingAt.FindStringSubmatch(f); m != nil {
		return "`" + m[1] + ":" + m[2] + "`", m[3]
	}
	if dir, rest, ok := strings.Cut(f, ": "); ok && !strings.ContainsAny(dir, " \t") {
		if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
			return "`" + dir + "`", rest
		}
	}
	return "", f
}

func cell(s string) string {
	return strings.ReplaceAll(shorten(s), "|", `\|`)
}

func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func lowerFirst(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[size:]
}

func count(n int, forms ...string) string {
	plural := forms[0] + "s"
	if len(forms) > 1 {
		plural = forms[1]
	}
	if n == 1 {
		return "1 " + forms[0]
	}
	return fmt.Sprintf("%d %s", n, plural)
}
