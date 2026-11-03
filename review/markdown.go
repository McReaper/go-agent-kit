package review

import (
	"fmt"
	"strings"

	"github.com/McReaper/go-agent-kit/spec"
)

// Marker starts every sheet, so a CI job can find and update its own
// comment on the pull request.
const Marker = "<!-- kit review -->"

func words(k spec.ChangeKind) string {
	if k == spec.Disabled {
		return "disabled: tagged `@wip`, it stops running"
	}
	return string(k)
}

// Markdown writes the sheet for a pull request comment.
func (s Sheet) Markdown() string {
	var b strings.Builder
	b.WriteString(Marker + "\n")
	if s.Empty() {
		b.WriteString("Nothing for you: this pull request merges by itself once its checks pass.\n")
		return b.String()
	}
	b.WriteString("## For you\n")
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
	b.WriteString("\nApproving the pull request approves these; touch one again and it needs approving again.\n")
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
