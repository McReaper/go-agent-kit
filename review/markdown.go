package review

import (
	"fmt"
	"strings"

	"github.com/McReaper/rumblestrip-go/spec"
)

// Marker starts every sheet, so a CI job can find and update its own
// comment on the pull request.
const Marker = "<!-- rumblestrip -->"

// StatusStart and StatusEnd hold the sheet's status line: the jobs that
// run after the sheet's, the check and the merge, rewrite what is between
// them as the pull request moves on.
const (
	StatusStart = "<!-- rumblestrip status -->"
	StatusEnd   = "<!-- /rumblestrip status -->"
)

// Repository is Rumblestrip's home, which the sheet's footer links to.
const Repository = "https://github.com/McReaper/rumblestrip-go"

func words(k spec.ChangeKind) string {
	if k == spec.Disabled {
		return "disabled: tagged `@wip`, it stops running"
	}
	return string(k)
}

// Status is the sheet's first line while the gates run: where the owner's
// approval stands.
func (s Sheet) Status() string {
	switch s.State {
	case Waiting:
		return "⏳ The gates are running. **Waiting for you:** read what is below, then approve the pull request."
	case Approved:
		return "⏳ The gates are running. You approved what is yours: Rumblestrip merges it once they pass."
	case NothingToApprove:
	}
	return "⏳ The gates are running. Nothing for you: Rumblestrip merges it once they pass."
}

// Markdown writes the sheet for a pull request comment; version is the
// Rumblestrip's release, empty when unknown.
func (s Sheet) Markdown(version string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n%s\n%s\n%s\n", Marker, StatusStart, s.Status(), StatusEnd)
	if !s.Empty() {
		b.WriteString("\n## For you\n")
	}
	if len(s.Spec) > 0 {
		b.WriteString("\n**The spec.** Read these scenarios in the pull request's files:\n\n")
		for _, c := range s.Spec {
			fmt.Fprintf(&b, "- %s, %s (`%s:%d`)\n", c.Name, words(c.Kind), c.File, c.Line)
		}
	}
	if len(s.Approval) > 0 {
		b.WriteString("\n**Also yours to approve:**\n\n")
		for _, it := range s.Approval {
			fmt.Fprintf(&b, "- %s (`%s`)\n", it.Text, where(it))
		}
	}
	if !s.Empty() {
		b.WriteString("\nApproving the pull request approves these; touch one again and it needs approving again.\n")
	}
	fmt.Fprintf(&b, "\n<sub>%s</sub>\n", strings.TrimSpace("[Rumblestrip]("+Repository+") "+version))
	return b.String()
}

func where(it Item) string {
	switch {
	case it.File == "":
		return "the change"
	case it.Line > 0:
		return fmt.Sprintf("%s:%d", it.File, it.Line)
	}
	return it.File
}
