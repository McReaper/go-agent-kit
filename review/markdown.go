package review

import (
	"fmt"
	"strings"
)

// Marker starts every sheet, so a CI job can find and update its own
// comment on the pull request.
const Marker = "<!-- kit review -->"

// Markdown writes the sheet for a pull request comment, empty sections
// left out.
func (s Sheet) Markdown() string {
	var b strings.Builder
	b.WriteString(Marker + "\n## Review sheet\n")
	n := b.Len()
	section(&b, "Needs your approval", "the merge check waits for a code owner's approval of the last commit", s.Approval)
	if !s.Tests.empty() {
		b.WriteString("\n### Test scenarios\n")
		list(&b, "Added", s.Tests.Added)
		list(&b, "Removed", s.Tests.Removed)
		list(&b, "Changed", s.Tests.Changed)
	}
	section(&b, "Contract", "golden snapshots of what the app's users see", s.Contract)
	section(&b, "Library API", "apidiff against the base", s.API)
	if len(s.Modules) > 0 {
		fmt.Fprintf(&b, "\n### New modules\n\n")
		for _, m := range s.Modules {
			fmt.Fprintf(&b, "- `%s`\n", m)
		}
	}
	section(&b, "Dependencies", "go.mod lines added", s.Deps)
	section(&b, "Shared code", "what other modules or the app build on", s.Shared)
	section(&b, "Rules", "CI, the agent's settings and instructions, decision records", s.Rules)
	section(&b, "Decisions", "`// WHY:` comments added", s.Why)
	if b.Len() == n {
		b.WriteString("\nNothing beyond the code itself: no protected change, no test scenario, no contract, dependency, shared code or rule touched.\n")
	}
	return b.String()
}

func section(b *strings.Builder, title, about string, items []Item) {
	if len(items) == 0 {
		return
	}
	fmt.Fprintf(b, "\n### %s\n\n_%s_\n\n", title, about)
	for _, it := range items {
		fmt.Fprintf(b, "- %s: %s\n", where(it), it.Text)
	}
}

func list(b *strings.Builder, title string, items []Item) {
	if len(items) == 0 {
		return
	}
	fmt.Fprintf(b, "\n%s:\n\n", title)
	for _, it := range items {
		fmt.Fprintf(b, "- %s (`%s`)\n", it.Text, it.File)
	}
}

func where(it Item) string {
	switch {
	case it.File == "":
		return "the change"
	case it.Line > 0:
		return fmt.Sprintf("`%s:%d`", it.File, it.Line)
	}
	return "`" + it.File + "`"
}
