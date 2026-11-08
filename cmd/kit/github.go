package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/McReaper/go-agent-kit/check"
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
// line, in the pull request's files, and in the run's one summary, which
// the Check job writes from each gate's section, where the owner reads
// what the gate wants and whose the finding is to fix.
func reportGitHub(out io.Writer, summaryPath string, results []check.Result) error {
	lines := annotations(results)
	if _, err := io.WriteString(out, strings.Join(lines[:min(len(lines), maxAnnotations)], "")); err != nil {
		return fmt.Errorf("annotate the findings: %w", err)
	}
	if summaryPath == "" {
		return nil
	}
	text := summary(results)
	if text == "" {
		return nil
	}
	return appendFile(summaryPath, text)
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
	return "::error file=" + escapeProperty(m[1]) + ",line=" + m[2] + ",title=" + escapeProperty(title) + "::" + escapeData(shorten(m[3])) + "\n", true
}

// WHY: a test's message can quote a whole program's output; the summary and
// the annotation say enough to know what broke, and the log keeps it all.
func shorten(s string) string {
	s = oneLine(s)
	if utf8.RuneCountInString(s) <= maxMessage {
		return s
	}
	return string([]rune(s)[:maxMessage]) + "…"
}

func escapeData(s string) string {
	return strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A").Replace(s)
}

func escapeProperty(s string) string {
	return strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A", ":", "%3A", ",", "%2C").Replace(s)
}

func summary(results []check.Result) string {
	var b strings.Builder
	for _, r := range results {
		switch {
		case r.Err != nil:
			fmt.Fprintf(&b, "### ❌ %s\n\nIt could not run: `%s`. That is the kit's or the runner's fault, not the change's: re-run the job.\n\n", r.Title, oneLine(r.Err.Error()))
		case len(r.Findings) > 0:
			fmt.Fprintf(&b, "### ❌ %s\n\nIt wants: %s It found %s, for the agent to fix:\n\n", r.Title, lowerFirst(r.Purpose), count(len(r.Findings), "problem"))
			writeRows(&b, r.Findings)
		case len(r.Notes) > 0:
			fmt.Fprintf(&b, "### ✅ %s\n\n%s\n\n", r.Title, r.Purpose)
		}
		if len(r.Notes) > 0 {
			fmt.Fprintf(&b, "%s yours to approve: the kit's comment on the pull request lists them.\n\n", count(len(r.Notes), "change is", "changes are"))
		}
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
		where, what := "", f
		if m := findingAt.FindStringSubmatch(f); m != nil {
			where, what = "`"+m[1]+":"+m[2]+"`", m[3]
		}
		fmt.Fprintf(b, "| %s | %s |\n", where, cell(what))
	}
	b.WriteString("\n")
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

func appendFile(path, text string) (err error) {
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return fmt.Errorf("write the job's summary: %w", err)
	}
	defer func() { err = errors.Join(err, root.Close()) }()
	f, err := root.OpenFile(filepath.Base(path), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("write the job's summary: %w", err)
	}
	defer func() { err = errors.Join(err, f.Close()) }()
	if _, err := f.WriteString(text); err != nil {
		return fmt.Errorf("write the job's summary: %w", err)
	}
	return nil
}
